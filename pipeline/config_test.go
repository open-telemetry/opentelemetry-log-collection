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

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/transformer/noop"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func newDummyOpConfig(dummyID string, dummyType string) *operator.Config {
	return &operator.Config{
		Builder: &noop.NoopOperatorConfig{
			TransformerConfig: helper.NewTransformerConfig(dummyID, dummyType),
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
	name            string
	ops             func() Config
	expectedOutputs []string
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
			[]string{
				"$.json_parser1",
			},
		},
		{
			"multi_op_rename",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				return ops
			},
			[]string{
				"$.json_parser1",
				"$.json_parser2",
				"$.json_parser3",
			},
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
			[]string{
				"$.json_parser1",
				"$.json_parser2",
				"$.copy",
				"$.copy1",
			},
		},
		{
			"unordered",
			func() Config {
				var ops Config
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("json_parser", "json_parser")})
				ops = append(ops, operator.Config{Builder: newDummyOpConfig("copy", "copy")})
				return ops
			},
			[]string{
				"$.copy",
				"$.json_parser1",
				"$.copy1",
			},
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
			[]string{
				"$.json_parser1",
				"$.json_parser2",
				"$.json_parser3",
				"$.json_parser4",
			},
		},
	}

	for _, tc := range cases {
		t.Run("UpdateOutputIDs/"+tc.name, func(t *testing.T) {
			bc := testutil.NewBuildContext(t)
			ops, err := tc.ops().BuildOperators(bc, nil)
			require.NoError(t, err)
			require.Equal(t, len(tc.expectedOutputs), len(ops)-1)
			for i := 0; i < len(ops)-1; i++ {
				require.Equal(t, tc.expectedOutputs[i], ops[i].GetOutputIDs()[0])
			}
		})
	}
}
