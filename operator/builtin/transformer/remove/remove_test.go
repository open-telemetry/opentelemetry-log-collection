package remove

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

func TestRemoveOperator(t *testing.T) {
	os.Setenv("TEST_REMOVE_PLUGIN_ENV", "foo")
	defer os.Unsetenv("TEST_REMOVE_PLUGIN_ENV")

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
