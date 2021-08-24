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

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

type dummyOpConfig struct {
	helper.WriterConfig
}

func (d *dummyOpConfig) ID() string {
	return d.OperatorID
}

func (d *dummyOpConfig) Type() string {
	return d.OperatorType
}

func (d *dummyOpConfig) SetID(id string) error {
	d.OperatorID = id
	return nil
}

func (c *dummyOpConfig) Build(bc operator.BuildContext) ([]operator.Operator, error) {
	// Namespace all the output IDs
	namespacedID := bc.PrependNamespace(c.ID())

	namespacedIDs := c.OutputIDs.WithNamespace(bc)
	if len(namespacedIDs) == 0 {
		namespacedIDs = bc.DefaultOutputIDs
	}

	dummy := dummyOperator{
		WriterOperator: helper.WriterOperator{
			OutputIDs: namespacedIDs,
			BasicOperator: helper.BasicOperator{
				OperatorType: c.OperatorType,
				OperatorID:   namespacedID,
			},
		},
	}

	return []operator.Operator{dummy}, nil
}

type dummyOperator struct {
	helper.WriterOperator
}

// ID returns the id of the operator.
func (d dummyOperator) ID() string { return d.OperatorID }

// Type returns the type of the operator.
func (d dummyOperator) Type() string { return d.OperatorType }

func (d dummyOperator) CanOutput() bool { return true }

func (d dummyOperator) CanProcess() bool { return true }

func (d dummyOperator) SetOutputs(operators []operator.Operator) error {
	outputOperators := make([]operator.Operator, 0)

	for _, operatorID := range d.OutputIDs {
		operator, ok := d.findOperator(operators, operatorID)
		if !ok {
			return fmt.Errorf("operator '%s' does not exist", operatorID)
		}

		if !operator.CanProcess() {
			return fmt.Errorf("operator '%s' can not process entries", operatorID)
		}

		outputOperators = append(outputOperators, operator)
	}

	d.OutputOperators = outputOperators
	return nil
}

func (w *dummyOperator) findOperator(operators []operator.Operator, operatorID string) (operator.Operator, bool) {
	for _, operator := range operators {
		if operator.ID() == operatorID {
			return operator, true
		}
	}
	return nil, false
}

func (d dummyOperator) Outputs() []operator.Operator { return nil }

func (d dummyOperator) GetOutputIDs() []string { return d.WriterOperator.OutputIDs }

func (d dummyOperator) SetOutputIDs(ids []string) { d.OutputIDs = ids }

func (d dummyOperator) Start(operator.Persister) error { return nil }

func (d dummyOperator) Stop() error { return nil }

// Process will process an entry from an operator.
func (d dummyOperator) Process(context.Context, *entry.Entry) error { return nil }

// Logger returns the operator's logger
func (d dummyOperator) Logger() *zap.SugaredLogger { return nil }

func newDummyOpConfig(dummyID string, dummyType string) *dummyOpConfig {
	return &dummyOpConfig{
		WriterConfig: helper.WriterConfig{
			BasicConfig: helper.BasicConfig{
				OperatorID:   dummyID,
				OperatorType: dummyType,
			},
		},
	}
}

func newDummyOp(dummyID string, dummyType string) dummyOperator {
	return dummyOperator{
		WriterOperator: helper.WriterOperator{
			BasicOperator: helper.BasicOperator{
				OperatorID:    dummyID,
				OperatorType:  dummyType,
				SugaredLogger: nil,
			},
		},
	}
}

type deduplicateTestCase struct {
	name        string
	ops         func() Config
	expectedOps Config
}

func TestDeduplicateIDs(t *testing.T) {
	cases := []deduplicateTestCase{
		{
			"one_op_rename",
			func() Config {
				var ops Config
				op1 := operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")}
				ops = append(ops, op1)
				op2 := operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")}
				ops = append(ops, op2)
				return ops
			},
			func() Config {
				var ops Config
				op1 := operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")}
				ops = append(ops, op1)
				op2 := operator.Config{Builder: newDummyOpConfig("json_parser1", "json_parser")}
				ops = append(ops, op2)
				return ops
			}(),
		},
		{
			"multi_op_rename",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})

				return ops
			},
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser2", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser4", "json_parser")})
				return ops
			}(),
		},
		{
			"different_ops",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})

				return ops
			},
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser2", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy1", "copy")})
				return ops
			}(),
		},
		{
			"unordered",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})

				return ops
			},
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy1", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser2", "json_parser")})
				return ops
			}(),
		},
		{
			"already_renamed",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})

				return ops
			},
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser2", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser4", "json_parser")})
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
}

type outputTestCase struct {
	name        string
	ops         func() Config
	expectedOps []operator.Operator
}

func TestUpdateOutputIDs(t *testing.T) {
	cases := []outputTestCase{
		{
			"one_op_rename",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				op1 := newDummyOp("$.json_parser", "json_parser")
				op1.OutputIDs = []string{"$.json_parser1"}
				ops = append(ops, op1)
				ops = append(ops, newDummyOp("$.json_parser1", "json_parser"))
				return ops
			}(),
		},
		{
			"multi_op_rename",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				op1 := newDummyOp("$.json_parser", "json_parser")
				op1.OutputIDs = []string{"$.json_parser1"}
				ops = append(ops, op1)
				op2 := newDummyOp("$.json_parser1", "json_parser")
				op2.OutputIDs = []string{"$.json_parser2"}
				ops = append(ops, op2)
				op3 := newDummyOp("$.json_parser2", "json_parser")
				op3.OutputIDs = []string{"$.json_parser3"}
				ops = append(ops, op3)
				op4 := newDummyOp("$.json_parser3", "json_parser")
				op4.OutputIDs = []string{"$.json_parser4"}
				ops = append(ops, op4)
				op5 := newDummyOp("$.json_parser4", "json_parser")
				ops = append(ops, op5)
				return ops
			}(),
		},
		{
			"different_ops",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				op1 := newDummyOp("$.json_parser", "json_parser")
				op1.OutputIDs = []string{"$.json_parser1"}
				ops = append(ops, op1)
				op2 := newDummyOp("$.json_parser1", "json_parser")
				op2.OutputIDs = []string{"$.json_parser2"}
				ops = append(ops, op2)
				op3 := newDummyOp("$.json_parser2", "json_parser")
				op3.OutputIDs = []string{"$.copy"}
				ops = append(ops, op3)
				op4 := newDummyOp("$.copy", "copy")
				op4.OutputIDs = []string{"$.copy1"}
				ops = append(ops, op4)
				op5 := newDummyOp("$.copy1", "copy")
				ops = append(ops, op5)
				return ops
			}(),
		},
		{
			"unordered",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				op1 := newDummyOp("$.json_parser", "json_parser")
				op1.OutputIDs = []string{"$.copy"}
				ops = append(ops, op1)
				op2 := newDummyOp("$.copy", "copy")
				op2.OutputIDs = []string{"$.json_parser1"}
				ops = append(ops, op2)
				op3 := newDummyOp("$.json_parser1", "json_parser")
				op3.OutputIDs = []string{"$.copy1"}
				ops = append(ops, op3)
				op4 := newDummyOp("$.copy1", "copy")
				op4.OutputIDs = []string{"$.json_parser2"}
				ops = append(ops, op4)
				op5 := newDummyOp("$.json_parser2", "json_parser")
				ops = append(ops, op5)
				return ops
			}(),
		},
		{
			"already_renamed",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				return ops
			},
			func() []operator.Operator {
				var ops []operator.Operator
				op1 := newDummyOp("$.json_parser", "json_parser")
				op1.OutputIDs = []string{"$.json_parser1"}
				ops = append(ops, op1)
				op2 := newDummyOp("$.json_parser1", "json_parser")
				op2.OutputIDs = []string{"$.json_parser2"}
				ops = append(ops, op2)
				op3 := newDummyOp("$.json_parser2", "json_parser")
				op3.OutputIDs = []string{"$.json_parser3"}
				ops = append(ops, op3)
				op4 := newDummyOp("$.json_parser3", "json_parser")
				op4.OutputIDs = []string{"$.json_parser4"}
				ops = append(ops, op4)
				op5 := newDummyOp("$.json_parser4", "json_parser")
				ops = append(ops, op5)
				return ops
			}(),
		},
	}

	for _, tc := range cases {
		t.Run("Deduplicate/"+tc.name, func(t *testing.T) {
			opsConfig := tc.ops()
			bc := testutil.NewBuildContext(t)
			ops, err := opsConfig.BuildOperators(bc, nil)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOps, ops)
			require.Equal(t, ops, tc.expectedOps)
		})
	}
}
