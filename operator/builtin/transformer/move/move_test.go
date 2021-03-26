package move

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

func TestMoveOperator(t *testing.T) {
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
		addOp     MoveOperator
		input     *entry.Entry
		output    *entry.Entry
		expectErr bool
	}{
		{
			name: "MoveValue",
			addOp: MoveOperator{
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
			name: "MoveNest",
			addOp: MoveOperator{
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
			cfg.To = tc.addOp.To
			cfg.From = tc.addOp.From
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
