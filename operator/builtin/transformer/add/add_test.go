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
package add

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

type testCase struct {
	name   string
	op     *AddOperatorConfig
	input  func() *entry.Entry
	output func() *entry.Entry
}

func TestAddGoldenConfig(t *testing.T) {
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

	cases := []testCase{
		{
			"add_value",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = "randomMessage"
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body.(map[string]interface{})["new"] = "randomMessage"
				return e
			},
		},
		{
			"add_expr",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = `EXPR($.key + "_suffix")`
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body.(map[string]interface{})["new"] = "val_suffix"
				return e
			},
		},
		{
			"add_nest",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = map[interface{}]interface{}{
					"nest": map[interface{}]interface{}{
						"key": "val",
					},
				}
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Body = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"new": map[interface{}]interface{}{
						"nest": map[interface{}]interface{}{
							"key": "val",
						},
					},
				}
				return e
			},
		},
		{
			"add_attribute",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewAttributeField("new")
				cfg.Value = "newVal"
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"new": "newVal"}
				return e
			},
		},
		{
			"add_resource",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = "newVal"
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "newVal"}
				return e
			},
		},
		{
			"add_resource_expr",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = `EXPR($.key + "_suffix")`
				return cfg
			}(),
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "val_suffix"}
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

			add := op.(*AddOperator)
			fake := testutil.NewFakeOutput(t)
			add.SetOutputs([]operator.Operator{fake})
			val := tc.input()
			err = add.Process(context.Background(), val)
			require.NoError(t, err)
			fake.ExpectEntry(t, tc.output())
		})
	}
}

func configFromFileViaYaml(file string) (*AddOperatorConfig, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}

	config := defaultCfg()
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("failed to read config file as yaml: %s", err)
	}

	return config, nil
}

func configFromFileViaMapstructure(file string, result *AddOperatorConfig) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not find config file: %s", err)
	}

	raw := map[string]interface{}{}

	if err := yaml.Unmarshal(bytes, raw); err != nil {
		return fmt.Errorf("failed to read data from yaml: %s", err)
	}

	err = helper.UnmarshalMapstructure(raw, result)
	return err
}

func defaultCfg() *AddOperatorConfig {
	return NewAddOperatorConfig("remove")
}
