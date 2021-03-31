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
	"fmt"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("copy", func() operator.Builder { return NewRemoveOperatorConfig("") })
}

// NewRemoveOperatorConfig creates a new restructure operator config with default values
func NewRemoveOperatorConfig(operatorID string) *RemoveOperatorConfig {
	return &RemoveOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "copy"),
	}
}

// RemoveOperatorConfig is the configuration of a restructure operator
type RemoveOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	From                     entry.Field `mapstructure:"from"  json:"from" yaml:"from"`
	To                       entry.Field
}

func (c RemoveOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	copyOperator := &RemoveOperator{
		TransformerOperator: transformerOperator,
		Fields:              c.Fields,
	}

	return []operator.Operator{copyOperator}, nil
}

type RemoveOperator struct {
	helper.TransformerOperator
	Fields []entry.Field
}

// Process will process an entry with a restructure transformation.
func (p *RemoveOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *RemoveOperator) Transform(entry *entry.Entry) error {
	if p.Fields != nil {
		for _, field := range p.Fields {
			entry.Delete(field)
		}
	} else {
		return fmt.Errorf("copy: missing required field 'fields'")
	}
	return nil
}
