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
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/stretchr/testify/require"
)

type dummyOp helper.BasicConfig

func (d *dummyOp) ID() string {
	return d.OperatorID
}

func (d *dummyOp) Type() string {
	return d.OperatorType
}

func (d *dummyOp) SetID(id string) error {
	d.OperatorID = id
	return nil
}

func (d *dummyOp) Build(bc operator.BuildContext) ([]operator.Operator, error) {
	return nil, nil
}

func newDummyOp(dummyID string, dummyType string) *dummyOp {
	return &dummyOp{
		OperatorID:   dummyID,
		OperatorType: dummyType,
	}
}

type deduplicateTestCase struct {
	name        string
	ops         func() []operator.Config
	expectedOps []operator.Config
}

func TestDeduplicateIDs(t *testing.T) {
	cases := []deduplicateTestCase{
		{
			"one_op_rename",
			func() []operator.Config {
				var ops []operator.Config
				op1 := operator.Config{Builder: newDummyOp("json_parser", "json_parser")}
				ops = append(ops, op1)
				op2 := operator.Config{Builder: newDummyOp("json_parser", "json_parser")}
				ops = append(ops, op2)
				return ops
			},
			func() []operator.Config {
				var ops []operator.Config
				op1 := operator.Config{Builder: newDummyOp("json_parser", "json_parser")}
				ops = append(ops, op1)
				op2 := operator.Config{Builder: newDummyOp("json_parser1", "json_parser")}
				ops = append(ops, op2)
				return ops
			}(),
		},
		{
			"multi_op_rename",
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})

				return ops
			},
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser2", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser4", "json_parser")})
				return ops
			}(),
		},
		{
			"different_ops",
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy", "copy")})

				return ops
			},
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser2", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy1", "copy")})
				return ops
			}(),
		},
		{
			"unordered",
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})

				return ops
			},
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("copy1", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser2", "json_parser")})
				return ops
			}(),
		},
		{
			"already_renamed",
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})

				return ops
			},
			func() []operator.Config {
				var ops []operator.Config
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser1", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser2", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser3", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOp("json_parser4", "json_parser")})
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
