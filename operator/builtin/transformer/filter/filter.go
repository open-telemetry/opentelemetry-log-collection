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

package filter

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("filter", func() operator.Builder { return NewFilterOperatorConfig("") })
}

// NewFilterOperatorConfig creates a filter operator config with default values
func NewFilterOperatorConfig(operatorID string) *FilterOperatorConfig {
	return &FilterOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "filter"),
		DropRatio:         1,
	}
}

// FilterOperatorConfig is the configuration of a filter operator
type FilterOperatorConfig struct {
	helper.TransformerConfig `yaml:",inline"`
	Expression               string  `json:"expr"   yaml:"expr"`
	DropRatio                float64 `json:"drop_ratio"   yaml:"drop_ratio"`
}

// Build will build a filter operator from the supplied configuration
func (c FilterOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformer, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	compiledExpression, err := expr.Compile(c.Expression, expr.AsBool(), expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression '%s': %w", c.Expression, err)
	}

	if c.DropRatio < 0.0 || c.DropRatio > 1.0 {
		return nil, fmt.Errorf("drop_ratio must be a number between 0 and 1")
	}

	filterOperator := &FilterOperator{
		TransformerOperator: transformer,
		expression:          compiledExpression,
		dropRatio:           c.DropRatio,
	}

	return []operator.Operator{filterOperator}, nil
}

// FilterOperator is an operator that filters entries based on matching expressions
type FilterOperator struct {
	helper.TransformerOperator
	expression *vm.Program
	dropRatio  float64
}

// Process will drop incoming entries that match the filter expression
func (f *FilterOperator) Process(ctx context.Context, entry *entry.Entry) error {
	env := helper.GetExprEnv(entry)
	defer helper.PutExprEnv(env)

	matches, err := vm.Run(f.expression, env)
	if err != nil {
		f.Errorf("Running expressing returned an error", zap.Error(err))
		return nil
	}

	filtered, ok := matches.(bool)
	if !ok {
		f.Errorf("Expression did not compile as a boolean")
		return nil
	}

	if !filtered || rand.Float64() > f.dropRatio {
		f.Write(ctx, entry)
	}

	return nil
}
