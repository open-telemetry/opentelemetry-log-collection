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
	"context"
	"fmt"
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/stretchr/testify/require"
)

type dummyOp struct {
	helper.BasicOperator
	OutputIDs []string
	OutputOps []operator.Operator
}

func newDummyOp(dummyID string, dummyType string) *dummyOp {
	return &dummyOp{
		BasicOperator: helper.BasicOperator{
			OperatorID:   dummyID,
			OperatorType: dummyType,
		},
	}
}

// SetOutputs will set the outputs of the operator.
func (w *dummyOp) SetOutputs(operators []operator.Operator) error {
	outputOperators := make([]operator.Operator, 0)

	for _, operatorID := range w.OutputIDs {
		operator, ok := w.findOperator(operators, operatorID)
		if !ok {
			return fmt.Errorf("operator '%s' does not exist", operatorID)
		}

		if !operator.CanProcess() {
			return fmt.Errorf("operator '%s' can not process entries", operatorID)
		}

		outputOperators = append(outputOperators, operator)
	}

	w.OutputOps = outputOperators
	return nil
}

// FindOperator will find an operator matching the supplied id.
func (w *dummyOp) findOperator(operators []operator.Operator, operatorID string) (operator.Operator, bool) {
	for _, operator := range operators {
		if operator.ID() == operatorID {
			return operator, true
		}
	}
	return nil, false
}

func (d dummyOp) CanOutput() bool {
	return true
}
func (d dummyOp) CanProcess() bool {
	return true
}

func (d dummyOp) Outputs() []operator.Operator {
	return d.OutputOps
}

func (d *dummyOp) Process(context.Context, *entry.Entry) error {
	return nil
}

type deduplicateTestCase struct {
	name        string
	ops         func() []operator.Operator
	expectedOps []operator.Operator
}

func TestDeduplicateIDs(t *testing.T) {
	cases := []deduplicateTestCase{
		{
			"one_op_rename",
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser1", "json_parser"))
				return ops
			}(),
		},
		{
			"multi_op_rename",
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))

				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser1", "json_parser"))
				ops = append(ops, newDummyOp("json_parser2", "json_parser"))
				ops = append(ops, newDummyOp("json_parser3", "json_parser"))
				ops = append(ops, newDummyOp("json_parser4", "json_parser"))
				return ops
			}(),
		},
		{
			"different_ops",
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("copy", "copy"))
				ops = append(ops, newDummyOp("copy", "copy"))

				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser1", "json_parser"))
				ops = append(ops, newDummyOp("json_parser2", "json_parser"))
				ops = append(ops, newDummyOp("copy", "copy"))
				ops = append(ops, newDummyOp("copy1", "copy"))
				return ops
			}(),
		},
		{
			"unordered",
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("copy", "copy"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("copy", "copy"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))

				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("copy", "copy"))
				ops = append(ops, newDummyOp("json_parser1", "json_parser"))
				ops = append(ops, newDummyOp("copy1", "copy"))
				ops = append(ops, newDummyOp("json_parser2", "json_parser"))
				return ops
			}(),
		},
		{
			"already_renamed",
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser3", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))

				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser1", "json_parser"))
				ops = append(ops, newDummyOp("json_parser2", "json_parser"))
				ops = append(ops, newDummyOp("json_parser3", "json_parser"))
				ops = append(ops, newDummyOp("json_parser4", "json_parser"))
				return ops
			}(),
		},
	}

	for _, tc := range cases {
		t.Run("Deduplicate/"+tc.name, func(t *testing.T) {
			ops := tc.ops()
			dedeplucateIDs(ops)
			require.Equal(t, ops, tc.expectedOps)
		})
	}

	cases = []deduplicateTestCase{
		{
			"one_op_rename",
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				ops = append(ops, newDummyOp("json_parser", "json_parser"))
				ops = append(ops, newDummyOp("json_parser1", "json_parser"))
				return ops
			}(),
		},
	}

	for _, tc := range cases {
		t.Run("DeduplicateAndFixOutputs/"+tc.name, func(t *testing.T) {
			ops := tc.ops()
			dedeplucateIDs(ops)
			require.Equal(t, ops, tc.expectedOps)
		})
	}
}
