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
package flatten

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
)

func TestFlattenOperator(t *testing.T) {
	os.Setenv("TEST_MOVE_PLUGIN_ENV", "foo")
	defer os.Unsetenv("TEST_MOVE_PLUGIN_ENV")

	newTestEntry := func() *entry.Entry {
		e := entry.New()
		e.Timestamp = time.Unix(1586632809, 0)
		e.Record = map[string]interface{}{
			"key": "val",
			"nested": map[string]interface{}{
				"nestedkey": "nestedval",
			},
		}
		return e
	}

	cases := []struct {
		name      string
		flattenOp FlattenOperator
		input     *entry.Entry
		output    *entry.Entry
		expectErr bool
	}{
		{
			name: "flatten_single_layer",
			flattenOp: FlattenOperator{
				Field: entry.RecordField{
					Keys: []string{"nested"},
				},
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key":       "val",
					"nestedkey": "nestedval",
				}
				return e
			}(),
		},
		{
			name: "flatten_double_layer",
			flattenOp: FlattenOperator{
				Field: entry.RecordField{
					Keys: []string{"nest1"},
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nest1": map[string]interface{}{
						"nest2": map[string]interface{}{
							"nestedkey": "nestedval",
						},
					},
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nest2": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewFlattenOperatorConfig("test")
			cfg.Field = tc.flattenOp.Field
			cfg.OutputIDs = []string{"fake"}
			ops, err := cfg.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			add := op.(*FlattenOperator)
			fake := testutil.NewFakeOutput(t)
			add.SetOutputs([]operator.Operator{fake})

			err = add.Process(context.Background(), tc.input)
			require.NoError(t, err)

			fake.ExpectEntry(t, tc.output)
		})
	}
}
