package move

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

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

type processTestCase struct {
	name      string
	expectErr bool
	op        *MoveOperatorConfig
	input     func() *entry.Entry
	output    func() *entry.Entry
}

func TestMoveProcess(t *testing.T) {
	newTestEntry := func() *entry.Entry {
		e := entry.New()
		e.Timestamp = time.Unix(1586632809, 0)
		e.Body = map[string]interface{}{
			"key": "val",
			"nested": map[string]interface{}{
				"nestedkey": "nestedval",
			},
		}
		return e
	}

	cases := []processTestCase{
		{
			"MoveBodyToBody",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("key")
				cfg.To = entry.NewBodyField("new")
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"new": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			},
		},
		{
			"MoveBodyToAttribute",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("key")
				cfg.To = entry.NewAttributeField("new")
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				e.Attributes = map[string]string{"new": "val"}
				return e
			},
		},
		{
			"MoveAttributeToBody",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewAttributeField("new")
				cfg.To = entry.NewBodyField("new")
				return cfg
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"new": "val"}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"new": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				e.Attributes = map[string]string{}
				return e
			},
		},
		{
			"MoveAttributeToResource",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewAttributeField("new")
				cfg.To = entry.NewResourceField("new")
				return cfg
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"new": "val"}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "val"}
				e.Attributes = map[string]string{}
				return e
			},
		},
		{
			"MoveResourceToAttribute",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewResourceField("new")
				cfg.To = entry.NewAttributeField("new")
				return cfg
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "val"}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{}
				e.Attributes = map[string]string{"new": "val"}
				return e
			},
		},
		{
			"MoveNest",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("nested")
				cfg.To = entry.NewBodyField("NewNested")
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"NewNested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			},
		},
		{
			"MoveFromNestedObj",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("nested", "nestedkey")
				cfg.To = entry.NewBodyField("unnestedkey")
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key":         "val",
					"nested":      map[string]interface{}{},
					"unnestedkey": "nestedval",
				}
				return e
			},
		},
		{
			"MoveToNestedObj",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("newnestedkey")
				cfg.To = entry.NewBodyField("nested", "newnestedkey")

				return cfg
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"newnestedkey": "nestedval",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey":    "nestedval",
						"newnestedkey": "nestedval",
					},
				}
				return e
			},
		},
		{
			"MoveDoubleNestedObj",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("nested", "nested2")
				cfg.To = entry.NewBodyField("nested2")
				return cfg
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
						"nested2": map[string]interface{}{
							"nestedkey": "nestedval",
						},
					},
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"nested2": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			},
		},
		{
			"MoveNestToResource",
			true,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("nested")
				cfg.To = entry.NewResourceField("NewNested")
				return cfg
			}(),
			newTestEntry,
			nil,
		},
		{
			"MoveNestToAttribute",
			true,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("nested")
				cfg.To = entry.NewAttributeField("NewNested")

				return cfg
			}(),
			newTestEntry,
			nil,
		},
		{
			"ReplaceBodyString",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewBodyField("nested")
				cfg.To = entry.NewBodyField()
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"nestedkey": "nestedval",
					"key":       "val",
				}
				return e
			},
		},
	}
	for _, tc := range cases {
		t.Run("BuildandProcess/"+tc.name, func(t *testing.T) {
			cfgFromMapstructure := tc.op
			cfgFromMapstructure.OutputIDs = []string{"fake"}
			cfgFromMapstructure.OnError = "drop"
			ops, err := cfgFromMapstructure.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			move := op.(*MoveOperator)
			fake := testutil.NewFakeOutput(t)
			move.SetOutputs([]operator.Operator{fake})
			val := tc.input()
			err = move.Process(context.Background(), val)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				fake.ExpectEntry(t, tc.output())
			}
		})
	}
}
