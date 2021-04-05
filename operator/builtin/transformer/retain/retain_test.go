package retain

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
	"fmt"
	"io/ioutil"
	"path"
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
	name      string
	expectErr bool
	input     func() *entry.Entry
	output    func() *entry.Entry
}

func TestRetainGoldenConfig(t *testing.T) {
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

	cases := []testCase{
		{
			"retain_single",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
				}
				return e
			},
		},
		{
			"retain_multi",
			false,
			func() *entry.Entry {
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
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested2": map[string]interface{}{
						"nestedkey": "nestedval",
					},
				}
				return e
			},
		},
		{
			"retain_single_attribute",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key": "val",
				}
				e.Record = nil
				return e
			},
		},
		{
			"retain_multi_attribute",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
					"key3": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				e.Record = nil
				return e
			},
		},
		{
			"retain_multi_attribute",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
					"key3": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				e.Record = nil
				return e
			},
		},
		{
			"retain_single_resource",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key": "val",
				}
				e.Record = nil
				return e
			},
		},
		{
			"retain_multi_resource",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
					"key3": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				e.Record = nil
				return e
			},
		},
		{
			"retain_one_of_each",
			false,
			func() *entry.Entry {
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
			},
			func() *entry.Entry {
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
			},
		},
	}
	for _, tc := range cases {
		t.Run("yaml/"+tc.name, func(t *testing.T) {
			cfgFromYaml, _ := configFromFileViaYaml(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			cfgFromYaml.OutputIDs = []string{"fake"}
			cfgFromYaml.OnError = "drop"
			ops, err := cfgFromYaml.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			remove := op.(*RetainOperator)
			fake := testutil.NewFakeOutput(t)
			remove.SetOutputs([]operator.Operator{fake})

			err = remove.Process(context.Background(), tc.input())
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				fake.ExpectEntry(t, tc.output())
			}
		})
		t.Run("mapstructure/"+tc.name, func(t *testing.T) {
			cfgFromMapstructure := defaultCfg()
			configFromFileViaMapstructure(
				path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)),
				cfgFromMapstructure,
			)
			cfgFromMapstructure.OutputIDs = []string{"fake"}
			cfgFromMapstructure.OnError = "drop"
			ops, err := cfgFromMapstructure.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			remove := op.(*RetainOperator)
			fake := testutil.NewFakeOutput(t)
			remove.SetOutputs([]operator.Operator{fake})
			val := tc.input()
			err = remove.Process(context.Background(), val)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				fake.ExpectEntry(t, tc.output())
			}
		})
	}
}

func configFromFileViaYaml(file string) (*RetainOperatorConfig, error) {
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

func configFromFileViaMapstructure(file string, result *RetainOperatorConfig) error {
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

func defaultCfg() *RetainOperatorConfig {
	return NewRetainOperatorConfig("remove")
}
