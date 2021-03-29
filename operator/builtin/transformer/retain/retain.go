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

package retain

import (
	"context"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("flatten", func() operator.Builder { return NewRetainOperatorConfig("") })
}

// NewRetainOperatorConfig creates a new restructure operator config with default values
func NewRetainOperatorConfig(operatorID string) *RetainOperatorConfig {
	return &RetainOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "move"),
	}
}

// RetainOperatorConfig is the configuration of a restructure operator
type RetainOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	Fields                   []entry.Field `json:"fields" yaml:"fields"`
}

func (c RetainOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	addOperator := &RetainOperator{
		TransformerOperator: transformerOperator,
		Fields:              c.Fields,
	}

	return []operator.Operator{addOperator}, nil
}

type RetainOperator struct {
	helper.TransformerOperator `mapstructure:",squash" yaml:",inline"`
	Fields                     []entry.Field
}

// Process will process an entry with a restructure transformation.
func (p *RetainOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *RetainOperator) Transform(e *entry.Entry) error {
	newEntry := entry.New()
	newEntry.Timestamp = e.Timestamp
	for _, field := range p.Fields {
		val, ok := e.Get(field)
		if !ok {
			continue
		}
		err := newEntry.Set(field, val)
		if err != nil {
			return err
		}
	}
	*e = *newEntry
	return nil
}
