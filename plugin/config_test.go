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

package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	yaml "gopkg.in/yaml.v2"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/transformer/noop"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/pipeline"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func TestGetRenderParams(t *testing.T) {
	cfg := Config{}
	cfg.OperatorID = "test"
	cfg.Parameters = map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
	}
	cfg.OutputIDs = []string{"out1", "out2"}

	params := cfg.getRenderParams(testutil.NewBuildContext(t))
	expected := map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
		"input":  "$.test",
		"id":     "test",
		"output": "[$.out1,$.out2]",
	}
	require.Equal(t, expected, params)
}

func TestPlugin(t *testing.T) {
	pluginContent := []byte(`
parameters:
pipeline:
  - id: {{ .input }}
    type: noop
    output: {{ .output }}
`)

	configContent := []byte(`
id: my_plugin_id
type: my_plugin
unused_param: test_unused
output: stdout
`)

	plugin, err := NewPlugin("my_plugin", pluginContent)
	require.NoError(t, err)

	operator.RegisterPlugin(plugin.ID, plugin.NewBuilder)

	var cfg operator.Config
	err = yaml.Unmarshal(configContent, &cfg)
	require.NoError(t, err)

	expected := operator.Config{
		Builder: &Config{
			WriterConfig: helper.WriterConfig{
				OutputIDs: []string{"stdout"},
				BasicConfig: helper.BasicConfig{
					OperatorID:   "my_plugin_id",
					OperatorType: "my_plugin",
				},
			},
			Plugin: plugin,
			Parameters: map[string]interface{}{
				"unused_param": "test_unused",
			},
		},
	}

	require.Equal(t, expected, cfg)

	operators, err := cfg.Build(testutil.NewBuildContext(t))
	require.NoError(t, err)
	require.Len(t, operators, 1)
	noop, ok := operators[0].(*noop.NoopOperator)
	require.True(t, ok)
	require.Equal(t, "send", noop.OnError)
	require.Equal(t, "$.my_plugin_id", noop.OperatorID)
	require.Equal(t, "noop", noop.OperatorType)
}

type PluginIDTestCase struct {
	Name          string
	PluginConfig  pipeline.Config
	ExpectedOpIDs []string
}

func TestPluginIDs(t *testing.T) {
	// TODO: ids shouldn't need to be specified once autogen IDs are implemented
	pluginContent := []byte(`
parameters:
pipeline:
  - id: noop
    type: noop
  - id: noop1
    type: noop
    output: {{ .output }}
`)
	pluginName := "my_plugin"
	pluginVar, err := NewPlugin(pluginName, pluginContent)
	require.NoError(t, err)
	operator.RegisterPlugin(pluginVar.ID, pluginVar.NewBuilder)

	// TODO: remove ID assignment
	pluginContent2 := []byte(`
parameters:
pipeline:
  - id: noop
    type: noop
  - id: noop1
    type: noop
`)
	secondPlugin := "secondPlugin"
	secondPluginVar, err := NewPlugin(secondPlugin, pluginContent2)
	require.NoError(t, err)
	operator.RegisterPlugin(secondPluginVar.ID, secondPluginVar.NewBuilder)

	cases := []PluginIDTestCase{
		{
			Name: "basic_plugin",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
				}
			}(),
			ExpectedOpIDs: []string{
				"$." + pluginName + ".noop",
				"$." + pluginName + ".noop1",
			},
		},
		{
			Name: "same_op_outside_plugin",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						// TODO: ID should be noop to start then auto gened to noop2
						Builder: noop.NewNoopOperatorConfig("noop2"),
					},
				}
			}(),
			ExpectedOpIDs: []string{
				"$." + pluginName + ".noop",
				"$." + pluginName + ".noop1",
				"$.noop2",
			},
		},
		{
			Name: "two_plugins_with_same_ops",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   secondPlugin,
									OperatorType: secondPlugin,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     secondPluginVar,
						},
					},
				}
			}(),
			ExpectedOpIDs: []string{
				"$." + pluginName + ".noop",
				"$." + pluginName + ".noop1",
				"$." + secondPlugin + ".noop",
				"$." + secondPlugin + ".noop1",
			},
		},
	}

	for _, tc := range cases {
		ops, err := tc.PluginConfig.BuildOperators(testutil.NewBuildContext(t))
		require.NoError(t, err)

		require.Len(t, ops, len(tc.ExpectedOpIDs))
		for i, op := range ops {
			require.Equal(t, tc.ExpectedOpIDs[i], op.ID())
		}
	}
}

type PluginOutputIDTestCase struct {
	Name          string
	PluginConfig  pipeline.Config
	ExpectedOpIDs map[string][]string
}

func TestPluginOutputIDs(t *testing.T) {
	// TODO: ids shouldn't need to be specified once autogen IDs are implemented
	pluginContent := []byte(`
parameters:
pipeline:
  - type: noop
  - id: noop1
    type: noop
    output: {{ .output }}
`)
	pluginName := "my_plugin"
	pluginVar, err := NewPlugin(pluginName, pluginContent)
	require.NoError(t, err)
	operator.RegisterPlugin(pluginVar.ID, pluginVar.NewBuilder)

	// TODO: remove ID assignment
	pluginContent2 := []byte(`
parameters:
pipeline:
  - type: noop
  - id: noop1
    type: noop
    output: {{ .output }}
`)
	secondPlugin := "secondPlugin"
	secondPluginVar, err := NewPlugin(secondPlugin, pluginContent2)
	require.NoError(t, err)
	operator.RegisterPlugin(secondPluginVar.ID, secondPluginVar.NewBuilder)

	cases := []PluginOutputIDTestCase{
		{
			Name: "same_op_outside_plugin",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: noop.NewNoopOperatorConfig("noop"),
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						// TODO: ID should be noop to start then auto gened to noop1
						Builder: noop.NewNoopOperatorConfig("noop1"),
					},
				}
			}(),
			ExpectedOpIDs: map[string][]string{
				"$.noop":                     {"$." + pluginName + ".noop"},
				"$." + pluginName + ".noop":  {"$." + pluginName + ".noop1"},
				"$." + pluginName + ".noop1": {"$.noop1"},
			},
		},
		{
			Name: "two_plugins_with_same_ops",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   secondPlugin,
									OperatorType: secondPlugin,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     secondPluginVar,
						},
					},
				}
			}(),
			ExpectedOpIDs: map[string][]string{
				"$." + pluginName + ".noop":   {"$." + pluginName + ".noop1"},
				"$." + pluginName + ".noop1":  {"$." + secondPlugin + ".noop"},
				"$." + secondPlugin + ".noop": {"$." + secondPlugin + ".noop1"},
			},
		},
		{
			Name: "two_plugins_specified_output",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
								OutputIDs: []string{"noop"},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   secondPlugin,
									OperatorType: secondPlugin,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     secondPluginVar,
						},
					},
					operator.Config{
						Builder: noop.NewNoopOperatorConfig("noop"),
					},
				}
			}(),
			ExpectedOpIDs: map[string][]string{
				"$." + pluginName + ".noop":    {"$." + pluginName + ".noop1"},
				"$." + pluginName + ".noop1":   {"$.noop"},
				"$." + secondPlugin + ".noop":  {"$." + secondPlugin + ".noop1"},
				"$." + secondPlugin + ".noop1": {"$.noop"},
			},
		},
		{
			Name: "two_plugins_output_to_non_sequential_plugin",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
								OutputIDs: []string{secondPlugin},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						Builder: noop.NewNoopOperatorConfig("noop5"),
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   secondPlugin,
									OperatorType: secondPlugin,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     secondPluginVar,
						},
					},
				}
			}(),
			ExpectedOpIDs: map[string][]string{
				"$." + pluginName + ".noop":   {"$." + pluginName + ".noop1"},
				"$." + pluginName + ".noop1":  {"$." + secondPlugin + ".noop"},
				"$.noop5":                     {"$." + secondPlugin + ".noop"},
				"$." + secondPlugin + ".noop": {"$." + secondPlugin + ".noop1"},
			},
		},
		{
			Name: "two_plugins_with_multiple_outside_ops",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: noop.NewNoopOperatorConfig("noop"),
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						// TODO: ID should be noop to start then auto gened to noop1
						Builder: noop.NewNoopOperatorConfig("noop1"),
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   secondPlugin,
									OperatorType: secondPlugin,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     secondPluginVar,
						},
					},
					operator.Config{
						// TODO: ID should be noop to start then auto gened to noop2
						Builder: noop.NewNoopOperatorConfig("noop2"),
					},
				}
			}(),
			ExpectedOpIDs: map[string][]string{
				"$.noop":                       {"$." + pluginName + ".noop"},
				"$." + pluginName + ".noop":    {"$." + pluginName + ".noop1"},
				"$." + pluginName + ".noop1":   {"$.noop1"},
				"$.noop1":                      {"$." + secondPlugin + ".noop"},
				"$." + secondPlugin + ".noop":  {"$." + secondPlugin + ".noop1"},
				"$." + secondPlugin + ".noop1": {"$.noop2"},
			},
		},
		{
			Name: "two_plugins_of_same_type",
			PluginConfig: func() []operator.Config {
				return pipeline.Config{
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName,
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     pluginVar,
						},
					},
					operator.Config{
						Builder: &Config{
							WriterConfig: helper.WriterConfig{
								BasicConfig: helper.BasicConfig{
									OperatorID:   pluginName + "1",
									OperatorType: pluginName,
								},
							},
							Parameters: map[string]interface{}{},
							Plugin:     secondPluginVar,
						},
					},
				}
			}(),
			ExpectedOpIDs: map[string][]string{
				"$." + pluginName + ".noop":  {"$." + pluginName + ".noop1"},
				"$." + pluginName + ".noop1": {"$." + pluginName + "1.noop"},
				"$." + pluginName + "1.noop": {"$." + pluginName + "1.noop1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ops, err := tc.PluginConfig.BuildOperators(testutil.NewBuildContext(t))
			require.NoError(t, err)

			for i, op := range ops {
				if i+1 < len(ops) {
					out := op.GetOutputIDs()
					t.Log("ID:" + op.ID())
					require.Equal(t, tc.ExpectedOpIDs[op.ID()], out)
				}
			}
		})
	}
}

func TestBuildRecursiveFails(t *testing.T) {
	pluginConfig1 := []byte(`
pipeline:
  - type: plugin2
`)

	pluginConfig2 := []byte(`
pipeline:
  - type: plugin1
`)

	plugin1, err := NewPlugin("plugin1", pluginConfig1)
	require.NoError(t, err)
	plugin2, err := NewPlugin("plugin2", pluginConfig2)
	require.NoError(t, err)

	t.Cleanup(func() { operator.DefaultRegistry = operator.NewRegistry() })
	operator.RegisterPlugin("plugin1", plugin1.NewBuilder)
	operator.RegisterPlugin("plugin2", plugin2.NewBuilder)

	pipelineConfig := []byte(`
- type: plugin1
`)

	var pipeline pipeline.Config
	err = yaml.Unmarshal(pipelineConfig, &pipeline)
	require.NoError(t, err)

	_, err = pipeline.BuildOperators(operator.NewBuildContext(zaptest.NewLogger(t).Sugar()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "reached max plugin depth")
}
