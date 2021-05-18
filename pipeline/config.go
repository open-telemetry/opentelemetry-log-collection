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

package pipeline

import (
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
)

// Config is the configuration of a pipeline.
type Config []operator.Config

// BuildOperators builds the operators from the list of configs into operators
func (c Config) BuildOperators(bc operator.BuildContext) ([]operator.Operator, error) {
	operators := make([]operator.Operator, 0, len(c))
	for _, builder := range c {
		op, err := builder.Build(bc)
		if err != nil {
			return nil, err
		}
		operators = append(operators, op...)
	}
	return operators, nil
}

// BuildPipeline will build a pipeline from the config.
func (c Config) BuildPipeline(bc operator.BuildContext, defaultOperator operator.Operator) (*DirectedPipeline, error) {
	if defaultOperator != nil {
		bc.DefaultOutputIDs = []string{defaultOperator.ID()}
	}

	operators, err := c.BuildOperators(bc)
	if err != nil {
		return nil, err
	}

	SetOutputIDs(operators)

	if defaultOperator != nil {
		operators = append(operators, defaultOperator)
	}

	return NewDirectedPipeline(operators)
}

func SetOutputIDs(operators []operator.Operator) error {
	for i, op := range operators {
		if len(op.GetOutputIDs()) == 0 && i+1 < len(operators) {
			err := op.SetOutputIDs([]string{operators[i+1].ID()})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getBuildContextWithDefaultOutput(configs []operator.Config, i int, bc operator.BuildContext) operator.BuildContext {
	if i+1 >= len(configs) {
		return bc
	}

	id := configs[i+1].ID()
	id = bc.PrependNamespace(id)
	return bc.WithDefaultOutputIDs([]string{id})
}
