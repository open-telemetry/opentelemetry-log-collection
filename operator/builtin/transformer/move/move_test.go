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
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
)

func TestMoveOperator(t *testing.T) {
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
		moveOp    MoveOperator
		input     *entry.Entry
		output    *entry.Entry
		expectErr bool
	}{
		{
			name: "MoveRecordToRecord",
			moveOp: MoveOperator{
				From: entry.NewRecordField("key"),
				To:   entry.NewRecordField("new"),
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"new": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
		},
		{
			name: "MoveRecordToAttribute",
			moveOp: MoveOperator{
				From: entry.NewRecordField("key"),
				To:   entry.NewAttributeField("new"),
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				e.Attributes = map[string]string{"new": "val"}
				return e
			}(),
		},
		{
			name: "MoveAttributeToRecord",
			moveOp: MoveOperator{
				From: entry.NewAttributeField("new"),
				To:   entry.NewRecordField("new"),
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"new": "val"}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"new": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				e.Attributes = map[string]string{}
				return e
			}(),
		},
		{
			name: "MoveAttributeToResource",
			moveOp: MoveOperator{
				From: entry.NewAttributeField("new"),
				To:   entry.NewResourceField("new"),
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"new": "val"}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "val"}
				e.Attributes = map[string]string{}
				return e
			}(),
		},
		{
			name: "MoveNest",
			moveOp: MoveOperator{
				From: entry.NewRecordField("nested"),
				To:   entry.NewRecordField("NewNested"),
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"NewNested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewMoveOperatorConfig("test")
			cfg.To = tc.moveOp.To
			cfg.From = tc.moveOp.From
			cfg.OutputIDs = []string{"fake"}
			ops, err := cfg.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			add := op.(*MoveOperator)
			fake := testutil.NewFakeOutput(t)
			add.SetOutputs([]operator.Operator{fake})

			err = add.Process(context.Background(), tc.input)
			require.NoError(t, err)

			fake.ExpectEntry(t, tc.output)
		})
	}
}
