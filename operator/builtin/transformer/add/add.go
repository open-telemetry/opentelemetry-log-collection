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

package add

import (
	"context"
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("add", func() operator.Builder { return NewAddOperatorConfig("") })
}

// NewAddOperatorConfig creates a new restructure operator config with default values
func NewAddOperatorConfig(operatorID string) *AddOperatorConfig {
	return &AddOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "add"),
	}
}

// AddOperatorConfig is the configuration of a restructure operator
type AddOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	Field                    entry.Field `mapstructure:"field" json:"field" yaml:"field"`
	Value                    interface{} `mapstructure:"value,omitempty" json:"value,omitempty" yaml:"value,omitempty"`
}

func (c AddOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	addOperator := &AddOperator{
		TransformerOperator: transformerOperator,
		Field:               c.Field,
	}
	_, ok := c.Value.(string)
	if ok && strings.Contains(c.Value.(string), "EXPR(") {
		exprStr := strings.TrimPrefix(c.Value.(string), "EXPR(")
		exprStr = strings.TrimSuffix(exprStr, ")")
		compiled, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression '%s': %w", c.IfExpr, err)
		}
		addOperator.program = compiled
	} else {
		addOperator.Value = c.Value
	}

	return []operator.Operator{addOperator}, nil
}

type AddOperator struct {
	helper.TransformerOperator

	Field   entry.Field
	Value   interface{}
	program *vm.Program
}

// Process will process an entry with a restructure transformation.
func (p *AddOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *AddOperator) Transform(entry *entry.Entry) error {
	if p.Value != nil {
		return entry.Set(p.Field, p.Value)
	} else if p.program != nil {
		env := helper.GetExprEnv(entry)
		defer helper.PutExprEnv(env)

		result, err := vm.Run(p.program, env)
		if err != nil {
			return fmt.Errorf("evaluate value_expr: %s", err)
		}
		return entry.Set(p.Field, result)
	}
	return fmt.Errorf("add: missing required field 'Value'")
}
