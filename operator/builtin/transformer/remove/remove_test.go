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

package remove

import (
	"context"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
)

func TestRemoveOperator(t *testing.T) {
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
		name        string
		removeItems []entry.Field
		input       *entry.Entry
		output      *entry.Entry
		expectErr   bool
	}{
		{
			name: "Remove_one",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewRecordField("nested"))
				return fields
			}(),
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
				}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_multi",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewRecordField("nested"))
				fields = append(fields, entry.NewRecordField("key"))
				return fields
			}(),
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_empty_value",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewRecordField(""))
				return fields
			}(),
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
			expectErr: false,
		},
		{
			name:        "Remove_nil_value",
			removeItems: nil,
			input:       newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_incorrect_key",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewRecordField("asdasd"))
				return fields
			}(),
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_special",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewRecordField("%$#"))
				return fields
			}(),
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"%$#": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_single_Attribute",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewAttributeField("key"))
				return fields
			}(),
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_nulti_Attribute",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewAttributeField("key1"))
				fields = append(fields, entry.NewAttributeField("key2"))

				return fields
			}(),
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_single_Resource",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewResourceField("key"))
				return fields
			}(),
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Remove_nulti_Resource",
			removeItems: func() []entry.Field {
				var fields []entry.Field
				fields = append(fields, entry.NewResourceField("key1"))
				fields = append(fields, entry.NewResourceField("key2"))

				return fields
			}(),
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{}
				return e
			}(),
			expectErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewRemoveOperatorConfig("test")
			cfg.OutputIDs = []string{"fake"}

			ops, err := cfg.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			remove := op.(*RemoveOperator)
			remove.Fields = tc.removeItems
			fake := testutil.NewFakeOutput(t)
			remove.SetOutputs([]operator.Operator{fake})

			err = remove.Process(context.Background(), tc.input)
			if tc.expectErr {
				require.Error(t, err)
			}
			require.NoError(t, err)

			fake.ExpectEntry(t, tc.output)
		})
	}
}
