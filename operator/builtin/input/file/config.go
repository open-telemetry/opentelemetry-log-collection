// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"fmt"
	"time"

	"github.com/bmatcuk/doublestar/v3"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("file_input", func() operator.Builder { return NewInputConfig("") })
}

const (
	defaultMaxLogSize         = 1024 * 1024
	defaultMaxConcurrentFiles = 1024
)

// NewInputConfig creates a new input config with default values
func NewInputConfig(operatorID string) *InputConfig {
	return &InputConfig{
		InputConfig:             helper.NewInputConfig(operatorID, "file_input"),
		PollInterval:            helper.Duration{Duration: 200 * time.Millisecond},
		IncludeFileName:         true,
		IncludeFilePath:         false,
		IncludeFileNameResolved: false,
		IncludeFilePathResolved: false,
		StartAt:                 "end",
		FingerprintSize:         defaultFingerprintSize,
		MaxLogSize:              defaultMaxLogSize,
		MaxConcurrentFiles:      defaultMaxConcurrentFiles,
		Encoding:                helper.NewEncodingConfig(),
	}
}

// InputConfig is the configuration of a file input operator
type InputConfig struct {
	helper.InputConfig `mapstructure:",squash" yaml:",inline"`

	Include []string `mapstructure:"include,omitempty" json:"include,omitempty" yaml:"include,omitempty"`
	Exclude []string `mapstructure:"exclude,omitempty" json:"exclude,omitempty" yaml:"exclude,omitempty"`

	PollInterval            helper.Duration        `mapstructure:"poll_interval,omitempty"                  json:"poll_interval,omitempty"                 yaml:"poll_interval,omitempty"`
	Multiline               helper.MultilineConfig `mapstructure:"multiline,omitempty"                      json:"multiline,omitempty"                     yaml:"multiline,omitempty"`
	IncludeFileName         bool                   `mapstructure:"include_file_name,omitempty"              json:"include_file_name,omitempty"             yaml:"include_file_name,omitempty"`
	IncludeFilePath         bool                   `mapstructure:"include_file_path,omitempty"              json:"include_file_path,omitempty"             yaml:"include_file_path,omitempty"`
	IncludeFileNameResolved bool                   `mapstructure:"include_file_name_resolved,omitempty"     json:"include_file_name_resolved,omitempty"    yaml:"include_file_name_resolved,omitempty"`
	IncludeFilePathResolved bool                   `mapstructure:"include_file_path_resolved,omitempty"     json:"include_file_path_resolved,omitempty"    yaml:"include_file_path_resolved,omitempty"`
	StartAt                 string                 `mapstructure:"start_at,omitempty"                       json:"start_at,omitempty"                      yaml:"start_at,omitempty"`
	FingerprintSize         helper.ByteSize        `mapstructure:"fingerprint_size,omitempty"               json:"fingerprint_size,omitempty"              yaml:"fingerprint_size,omitempty"`
	MaxLogSize              helper.ByteSize        `mapstructure:"max_log_size,omitempty"                   json:"max_log_size,omitempty"                  yaml:"max_log_size,omitempty"`
	MaxConcurrentFiles      int                    `mapstructure:"max_concurrent_files,omitempty"           json:"max_concurrent_files,omitempty"          yaml:"max_concurrent_files,omitempty"`
	Encoding                helper.EncodingConfig  `mapstructure:",squash,omitempty"                        json:",inline,omitempty"                       yaml:",inline,omitempty"`
}

// Build will build a file input operator from the supplied configuration
func (c InputConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	if len(c.Include) == 0 {
		return nil, fmt.Errorf("required argument `include` is empty")
	}

	// Ensure includes can be parsed as globs
	for _, include := range c.Include {
		_, err := doublestar.PathMatch(include, "matchstring")
		if err != nil {
			return nil, fmt.Errorf("parse include glob: %s", err)
		}
	}

	// Ensure excludes can be parsed as globs
	for _, exclude := range c.Exclude {
		_, err := doublestar.PathMatch(exclude, "matchstring")
		if err != nil {
			return nil, fmt.Errorf("parse exclude glob: %s", err)
		}
	}

	if c.MaxLogSize <= 0 {
		return nil, fmt.Errorf("`max_log_size` must be positive")
	}

	if c.MaxConcurrentFiles <= 1 {
		return nil, fmt.Errorf("`max_concurrent_files` must be greater than 1")
	}

	if c.FingerprintSize == 0 {
		c.FingerprintSize = defaultFingerprintSize
	} else if c.FingerprintSize < minFingerprintSize {
		return nil, fmt.Errorf("`fingerprint_size` must be at least %d bytes", minFingerprintSize)
	}

	encoding, err := c.Encoding.Build(context)
	if err != nil {
		return nil, err
	}

	splitFunc, err := c.Multiline.Build(encoding.Encoding, false, helper.NewForceFlush())
	if err != nil {
		return nil, err
	}

	var startAtBeginning bool
	switch c.StartAt {
	case "beginning":
		startAtBeginning = true
	case "end":
		startAtBeginning = false
	default:
		return nil, fmt.Errorf("invalid start_at location '%s'", c.StartAt)
	}

	fileNameField := entry.NewNilField()
	if c.IncludeFileName {
		fileNameField = entry.NewAttributeField("file.name")
	}

	filePathField := entry.NewNilField()
	if c.IncludeFilePath {
		filePathField = entry.NewAttributeField("file.path")
	}

	fileNameResolvedField := entry.NewNilField()
	if c.IncludeFileNameResolved {
		fileNameResolvedField = entry.NewAttributeField("file.name.resolved")
	}

	filePathResolvedField := entry.NewNilField()
	if c.IncludeFilePathResolved {
		filePathResolvedField = entry.NewAttributeField("file.path.resolved")
	}

	op := &InputOperator{
		InputOperator:         inputOperator,
		Include:               c.Include,
		Exclude:               c.Exclude,
		SplitFunc:             splitFunc,
		Multiline:             c.Multiline,
		PollInterval:          c.PollInterval.Raw(),
		FilePathField:         filePathField,
		FileNameField:         fileNameField,
		FilePathResolvedField: filePathResolvedField,
		FileNameResolvedField: fileNameResolvedField,
		startAtBeginning:      startAtBeginning,
		queuedMatches:         make([]string, 0),
		encoding:              encoding,
		firstCheck:            true,
		cancel:                func() {},
		knownFiles:            make([]*Reader, 0, 10),
		fingerprintSize:       int(c.FingerprintSize),
		MaxLogSize:            int(c.MaxLogSize),
		MaxConcurrentFiles:    c.MaxConcurrentFiles,
		SeenPaths:             make(map[string]struct{}, 100),
	}

	return []operator.Operator{op}, nil
}
