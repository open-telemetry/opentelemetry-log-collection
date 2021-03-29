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

package move

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("move", func() operator.Builder { return NewMoveOperatorConfig("") })
}

// NewMoveOperatorConfig creates a new restructure operator config with default values
func NewMoveOperatorConfig(operatorID string) *MoveOperatorConfig {
	return &MoveOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "move"),
	}
}

// MoveOperatorConfig is the configuration of a restructure operator
type MoveOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	From                     entry.Field `json:"from" yaml:"from,flow"`
	To                       entry.Field `json:"to" yaml:"to,flow"`
}

func (c MoveOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	addOperator := &MoveOperator{
		TransformerOperator: transformerOperator,
		From:                c.From,
		To:                  c.To,
	}

	return []operator.Operator{addOperator}, nil
}

type MoveOperator struct {
	helper.TransformerOperator `mapstructure:",squash" yaml:",inline"`
	From                       entry.Field `json:"from" yaml:"from,flow"`
	To                         entry.Field `json:"to" yaml:"to,flow"`
}

// Process will process an entry with a restructure transformation.
func (p *MoveOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *MoveOperator) Transform(entry *entry.Entry) error {
	val, ok := entry.Delete(p.From)
	if !ok {
		return fmt.Errorf("apply move: field %s does not exist on record", p.From)
	}

	return entry.Set(p.To, val)
}
