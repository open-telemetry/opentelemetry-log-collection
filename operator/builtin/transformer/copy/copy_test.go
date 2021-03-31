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

package copy

import (
	"context"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
)

func TestCopyOperator(t *testing.T) {
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
		copyOp    *CopyOperator
		input     *entry.Entry
		output    *entry.Entry
		expectErr bool
	}{
		{
			name: "Copy_Record",
			copyOp: &CopyOperator{
				From: entry.NewRecordField("key"),
				To:   entry.NewRecordField("key2"),
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"key2": "val",
				}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Copy_record_to_attribute",
			copyOp: &CopyOperator{
				From: entry.NewRecordField("key"),
				To:   entry.NewAttributeField("key2"),
			},
			input: newTestEntry(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				e.Attributes = map[string]string{"key2": "val"}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Copy_attribute_to_record",
			copyOp: &CopyOperator{
				From: entry.NewAttributeField("key2"),
				To:   entry.NewRecordField("key3"),
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"key2": "val"}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"key3": "val",
				}
				e.Attributes = map[string]string{"key2": "val"}
				return e
			}(),
			expectErr: false,
		},
		{
			name: "Copy_attribute_to_resource",
			copyOp: &CopyOperator{
				From: entry.NewAttributeField("key2"),
				To:   entry.NewResourceField("key3"),
			},
			input: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"key2": "val"}
				return e
			}(),
			output: func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"key2": "val"}
				e.Resource = map[string]string{"key3": "val"}
				return e
			}(),
			expectErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewCopyOperatorConfig("test")
			cfg.OutputIDs = []string{"fake"}

			ops, err := cfg.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			remove := op.(*CopyOperator)
			remove.From = tc.copyOp.From
			remove.To = tc.copyOp.To
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
