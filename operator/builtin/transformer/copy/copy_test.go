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
	input     *entry.Entry
	output    *entry.Entry
}

func TestCopyGoldenConfig(t *testing.T) {
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
			"record_to_record",
			false,
			newTestEntry(),
			func() *entry.Entry {
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
		},
		{
			"record_to_attribute",
			false,
			newTestEntry(),
			func() *entry.Entry {
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
		},
		{
			"attribute_to_record",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"key": "val"}
				return e
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
					"key": "val",
					"nested": map[string]interface{}{
						"nestedkey": "nestedval",
					},
					"key2": "val",
				}
				e.Attributes = map[string]string{"key": "val"}
				return e
			}(),
		},
		{
			"attribute_to_resource",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"key": "val"}
				return e
			}(),
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"key": "val"}
				e.Resource = map[string]string{"key2": "val"}
				return e
			}(),
		},
	}

	for _, tc := range cases {
		t.Run("yaml/"+tc.name, func(t *testing.T) {
			cfgFromYaml, yamlErr := configFromFileViaYaml(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			if tc.expectErr {
				require.Error(t, yamlErr)
			} else {
				cfgFromYaml.OutputIDs = []string{"fake"}
				ops, err := cfgFromYaml.Build(testutil.NewBuildContext(t))
				require.NoError(t, err)
				op := ops[0]

				remove := op.(*CopyOperator)
				fake := testutil.NewFakeOutput(t)
				remove.SetOutputs([]operator.Operator{fake})

				err = remove.Process(context.Background(), tc.input)
				if tc.expectErr {
					require.Error(t, err)
				}
				require.NoError(t, err)

				fake.ExpectEntry(t, tc.output)
			}
		})
		t.Run("mapstructure/"+tc.name, func(t *testing.T) {
			cfgFromMapstructure := defaultCfg()
			mapErr := configFromFileViaMapstructure(
				path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)),
				cfgFromMapstructure,
			)
			if tc.expectErr {
				require.Error(t, mapErr)
			} else {
				cfgFromMapstructure.OutputIDs = []string{"fake"}
				ops, err := cfgFromMapstructure.Build(testutil.NewBuildContext(t))
				require.NoError(t, err)
				op := ops[0]

				remove := op.(*CopyOperator)
				fake := testutil.NewFakeOutput(t)
				remove.SetOutputs([]operator.Operator{fake})

				err = remove.Process(context.Background(), tc.input)
				if tc.expectErr {
					require.Error(t, err)
				}
				require.NoError(t, err)

				fake.ExpectEntry(t, tc.output)
			}
		})
	}
}

func configFromFileViaYaml(file string) (*CopyOperatorConfig, error) {
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

func configFromFileViaMapstructure(file string, result *CopyOperatorConfig) error {
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

func defaultCfg() *CopyOperatorConfig {
	return NewCopyOperatorConfig("copy")
}
