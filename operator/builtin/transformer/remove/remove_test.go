package remove

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
	yaml "gopkg.in/yaml.v2"
)

type testCase struct {
	name      string
	expectErr bool
	input     func() *entry.Entry
	output    func() *entry.Entry
}

func TestRemoveGoldenConfig(t *testing.T) {
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
			"remove_one",
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
			"remove_multi",
			false,
			newTestEntry,
			func() *entry.Entry {
				e := newTestEntry()
				e.Record = map[string]interface{}{}
				return e
			},
		},
		{
			"remove_single_attribute",
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
				e.Attributes = map[string]string{}
				return e
			},
		},
		{
			"remove_multi_attribute",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Attributes = map[string]string{}
				return e
			},
		},
		{
			"remove_single_resource",
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
				e.Resource = map[string]string{}
				return e
			},
		},
		{
			"remove_multi_resource",
			false,
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{
					"key1": "val",
					"key2": "val",
				}
				return e
			},
			func() *entry.Entry {
				e := newTestEntry()
				e.Resource = map[string]string{}
				return e
			},
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

				remove := op.(*RemoveOperator)
				fake := testutil.NewFakeOutput(t)
				remove.SetOutputs([]operator.Operator{fake})

				err = remove.Process(context.Background(), tc.input())
				if tc.expectErr {
					require.Error(t, err)
				}
				require.NoError(t, err)

				fake.ExpectEntry(t, tc.output())
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

				remove := op.(*RemoveOperator)
				fake := testutil.NewFakeOutput(t)
				remove.SetOutputs([]operator.Operator{fake})

				err = remove.Process(context.Background(), tc.input())
				if tc.expectErr {
					require.Error(t, err)
				}
				require.NoError(t, err)

				fake.ExpectEntry(t, tc.output())
			}
		})
	}
}

func configFromFileViaYaml(file string) (*RemoveOperatorConfig, error) {
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

func configFromFileViaMapstructure(file string, result *RemoveOperatorConfig) error {
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

func defaultCfg() *RemoveOperatorConfig {
	return NewRemoveOperatorConfig("remove")
}
