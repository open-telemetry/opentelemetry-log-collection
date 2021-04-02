package add

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type testCase struct {
	name      string
	expectErr bool
	input     func() *entry.Entry
	output    func() *entry.Entry
}

func TestAddGoldenConfig(t *testing.T) {
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
			"add_value",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Record.(map[string]interface{})["new"] = "randomMessage"
				return e
			},
		},
		{
			"add_expr",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Record.(map[string]interface{})["new"] = "val_suffix"
				return e
			},
		},
		{
			"add_nest",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{
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
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{"new": "newVal"}
				return e
			},
		},
		{
			"add_resource",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "newVal"}
				return e
			},
		},
		{
			"add_resource_expr",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{"new": "val_suffix"}
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

			add := op.(*AddOperator)
			fake := testutil.NewFakeOutput(t)
			add.SetOutputs([]operator.Operator{fake})

			err = add.Process(context.Background(), tc.input())
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

			remove := op.(*AddOperator)
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
