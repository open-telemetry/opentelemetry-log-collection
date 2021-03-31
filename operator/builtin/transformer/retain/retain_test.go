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

package retain

import (
	"context"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
)

func TestOperator(t *testing.T) {
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
		retainOp  RetainOperator
		input     *entry.Entry
		output    *entry.Entry
		expectErr bool
	}{
		{
			name: "retain_single",
			retainOp: RetainOperator{
				Fields: []entry.Field{entry.NewRecordField("key")},
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
				}
				return e
			}(),
		},
		{
			name: "retain_multi",
			retainOp: RetainOperator{
				Fields: []entry.Field{
					entry.NewRecordField("key"),
					entry.NewRecordField("nested2"),
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"nested2": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested2": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
		},
		{
			name: "retain_single_attribute",
			retainOp: RetainOperator{
				Fields: []entry.Field{
					entry.NewAttributeField("key"),
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key": "val",
				}
				e.Record = nil
				return e
			}(),
		},
		{
			name: "retain_multi_attribute",
			retainOp: RetainOperator{
				Fields: []entry.Field{
					entry.NewAttributeField("key1"),
					entry.NewAttributeField("key2"),
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
					"key3": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				e.Record = nil
				return e
			}(),
		},
		{
			name: "retain_single_resource",
			retainOp: RetainOperator{
				Fields: []entry.Field{
					entry.NewResourceField("key"),
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key": "val",
				}
				e.Record = nil
				return e
			}(),
		},
		{
			name: "retain_multi_resource",
			retainOp: RetainOperator{
				Fields: []entry.Field{
					entry.NewResourceField("key1"),
					entry.NewResourceField("key2"),
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
					"key3": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				e.Record = nil
				return e
			}(),
		},
		{
			name: "retain_one_of_each",
			retainOp: RetainOperator{
				Fields: []entry.Field{
					entry.NewRecordField("key"),
					entry.NewResourceField("key1"),
					entry.NewAttributeField("key3"),
				},
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				e.Attributes = map[string]string{
					"key3": "val",
					"key4": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
				}
				e.Attributes = map[string]string{
					"key3": "val",
				}
				e.Record = map[string]interface{}{
					"key": "val",
				}
				return e
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewRetainOperatorConfig("test")
			cfg.Fields = tc.retainOp.Fields
			cfg.OutputIDs = []string{"fake"}
			ops, err := cfg.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			add := op.(*RetainOperator)
			fake := testutil.NewFakeOutput(t)
			add.SetOutputs([]operator.Operator{fake})

			err = add.Process(context.Background(), tc.input)
			require.NoError(t, err)

			fake.ExpectEntry(t, tc.output)
		})
	}
}
