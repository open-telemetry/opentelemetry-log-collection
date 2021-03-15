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

package severity

import (
	"context"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("severity_parser", func() operator.Builder { return NewSeverityParserConfig("") })
}

// NewSeverityParserConfig creates a new severity parser config with default values
func NewSeverityParserConfig(operatorID string) *SeverityParserConfig {
	return &SeverityParserConfig{
		TransformerConfig:    helper.NewTransformerConfig(operatorID, "severity_parser"),
		SeverityParserConfig: helper.NewSeverityParserConfig(),
	}
}

// SeverityParserConfig is the configuration of a severity parser operator.
type SeverityParserConfig struct {
	helper.TransformerConfig    `mapstructure:",squash" yaml:",inline"`
	helper.SeverityParserConfig `mapstructure:",omitempty,squash" yaml:",omitempty,inline"`
}

// Build will build a time parser operator.
func (c SeverityParserConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	severityParser, err := c.SeverityParserConfig.Build(context)
	if err != nil {
		return nil, err
	}

	severityOperator := &SeverityParserOperator{
		TransformerOperator: transformerOperator,
		SeverityParser:      severityParser,
	}

	return []operator.Operator{severityOperator}, nil
}

// SeverityParserOperator is an operator that parses time from a field to an entry.
type SeverityParserOperator struct {
	helper.TransformerOperator
	helper.SeverityParser
}

// Process will parse time from an entry.
func (p *SeverityParserOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Parse)
}
