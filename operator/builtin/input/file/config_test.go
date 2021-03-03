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
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type testCase struct {
	name      string
	expectErr bool
	expect    *InputConfig
}

func TestConfig(t *testing.T) {
	cases := []testCase{
		{
			"default",
			false,
			defaultCfg(),
		},
		{
			"id_custom",
			false,
			NewInputConfig("test_id"),
		},
		{
			"include_one",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				return cfg
			}(),
		},
		{
			"include_multi",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log", "two.log", "three.log")
				return cfg
			}(),
		},
		{
			"include_glob",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				return cfg
			}(),
		},
		{
			"include_inline",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "a.log", "b.log")
				return cfg
			}(),
		},
		{
			"include_invalid",
			true,
			func() *InputConfig {
				cfg := defaultCfg()
				return cfg
			}(),
		},
		{
			"exclude_one",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "one.log")
				return cfg
			}(),
		},
		{
			"exclude_multi",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "one.log", "two.log", "three.log")
				return cfg
			}(),
		},
		{
			"exclude_glob",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "not*.log")
				return cfg
			}(),
		},
		{
			"exclude_inline",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "a.log", "b.log")
				return cfg
			}(),
		},
		{
			"exclude_invalid",
			true,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				return cfg
			}(),
		},
		{
			"poll_interval_no_units",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Second)
				return cfg
			}(),
		},
		{
			"poll_interval_1s",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Second)
				return cfg
			}(),
		},
		{
			"poll_interval_1ms",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Millisecond)
				return cfg
			}(),
		},
		{
			"poll_interval_1000ms",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Second)
				return cfg
			}(),
		},
		{
			"fingerprint_size_no_units",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1000)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1kb_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1000)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1KB",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1000)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1kib_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1024)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1KiB",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1024)
				return cfg
			}(),
		},
		{
			"fingerprint_size_float",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1100)
				return cfg
			}(),
		},
		{
			"include_file_name_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFileName = true
				return cfg
			}(),
		},
		{
			"include_file_name_upper",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFileName = true
				return cfg
			}(),
		},
		{
			"include_file_name_on",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFileName = true
				return cfg
			}(),
		},
		{
			"include_file_name_yes",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFileName = true
				return cfg
			}(),
		},
		{
			"include_file_path_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFilePath = true
				return cfg
			}(),
		},
		{
			"include_file_path_upper",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFilePath = true
				return cfg
			}(),
		},
		{
			"include_file_path_on",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFilePath = true
				return cfg
			}(),
		},
		{
			"include_file_path_yes",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFilePath = true
				return cfg
			}(),
		},
		{
			"include_file_path_off",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFilePath = false
				return cfg
			}(),
		},
		{
			"include_file_path_no",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				cfg.IncludeFilePath = false
				return cfg
			}(),
		},
		{
			"include_file_path_nonbool",
			true,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				return cfg
			}(),
		},
		{
			"multiline_line_start_string",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				var newMulti *MultilineConfig
				newMulti = new(MultilineConfig)
				newMulti.LineStartPattern = "Start"
				cfg.Multiline = newMulti
				return cfg
			}(),
		},
		{
			"multiline_line_start_special",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				var newMulti *MultilineConfig
				newMulti = new(MultilineConfig)
				newMulti.LineStartPattern = "%"
				cfg.Multiline = newMulti
				return cfg
			}(),
		},
		{
			"multiline_line_end_string",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				var newMulti *MultilineConfig
				newMulti = new(MultilineConfig)
				newMulti.LineEndPattern = "Start"
				cfg.Multiline = newMulti
				return cfg
			}(),
		},
		{
			"multiline_line_end_special",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				var newMulti *MultilineConfig
				newMulti = new(MultilineConfig)
				newMulti.LineEndPattern = "%"
				cfg.Multiline = newMulti
				return cfg
			}(),
		},
		{
			"multiline_random",
			true,
			func() *InputConfig {
				cfg := defaultCfg()
				var newMulti *MultilineConfig
				newMulti = new(MultilineConfig)
				cfg.Multiline = newMulti
				return cfg
			}(),
		},
		{
			"start_at_string",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.StartAt = "beginning"
				return cfg
			}(),
		},
		{
			"start_at_empty",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.StartAt = ""
				return cfg
			}(),
		},
		{
			"max_concurrent_large",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.MaxConcurrentFiles = 9223372036854775807
				return cfg
			}(),
		},
		{
			"max_concurrent_empty",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.MaxConcurrentFiles = 0
				return cfg
			}(),
		},
		{
			"max_logs_mib_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.MaxLogSize = helper.ByteSize(1048576)
				return cfg
			}(),
		},
		{
			"max_logs_mib_upper",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.MaxLogSize = helper.ByteSize(1048576)
				return cfg
			}(),
		},
		{
			"max_logs_mb_upper",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.MaxLogSize = helper.ByteSize(1048576)
				return cfg
			}(),
		},
		{
			"max_logs_mb_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.MaxLogSize = helper.ByteSize(1048576)
				return cfg
			}(),
		},
		{
			"max_logs_invalid_unit",
			true,
			func() *InputConfig {
				cfg := defaultCfg()
				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFromYaml, err := configFromFileViaYaml(t, path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, cfgFromYaml)
			}
			// TODO cfgFromMapstructure, err := configFromFileViaYaml(t, path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			// TODO require.NoError(t, err)
			// TODO require.Equal(t, tc.expect, cfgFromMapstructure)
		})
	}
}

func configFromFileViaYaml(t *testing.T, file string) (*InputConfig, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}

	config := NewInputConfig("file_input")
	if err := yaml.UnmarshalStrict(bytes, config); err != nil {
		return nil, fmt.Errorf("failed to read config file as yaml: %s", err)
	}

	return config, nil
}

// TODO func configFromFileViaMapstructure(t *testing.T, file string) (*InputConfig, error) {}

func defaultCfg() *InputConfig {
	return NewInputConfig("file_input")
}
