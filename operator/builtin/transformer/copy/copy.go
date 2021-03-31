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

package copy

import (
	"context"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("copy", func() operator.Builder { return NewCopyOperatorConfig("") })
}

// NewCopyOperatorConfig creates a new restructure operator config with default values
func NewCopyOperatorConfig(operatorID string) *CopyOperatorConfig {
	return &CopyOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "copy"),
	}
}

// CopyOperatorConfig is the configuration of a restructure operator
type CopyOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	From                     entry.Field
	To                       entry.Field
}

func (c CopyOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	addOperator := &CopyOperator{
		TransformerOperator: transformerOperator,
		From:                c.From,
		To:                  c.To,
	}

	return []operator.Operator{addOperator}, nil
}

type CopyOperator struct {
	helper.TransformerOperator
	From entry.Field
	To   entry.Field
}

// Process will process an entry with a restructure transformation.
func (p *CopyOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *CopyOperator) Transform(e *entry.Entry) error {
	val, _ := p.From.Get(e)
	p.To.Set(e, val)
	return nil
}
