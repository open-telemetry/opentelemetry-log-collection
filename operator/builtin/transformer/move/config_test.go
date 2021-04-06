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
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

type configTestCase struct {
	name      string
	expectErr bool
	expect    *MoveOperatorConfig
}

func TestMoveGoldenConfig(t *testing.T) {
	cases := []configTestCase{
		{
			"MoveRecordToRecord",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("key")
				cfg.To = entry.NewRecordField("new")
				return cfg
			}(),
		},
		{
			"MoveRecordToAttribute",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("key")
				cfg.To = entry.NewAttributeField("new")
				return cfg
			}(),
		},
		{
			"MoveAttributeToRecord",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewAttributeField("new")
				cfg.To = entry.NewRecordField("new")
				return cfg
			}(),
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
		},
		{
			"MoveNest",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("nested")
				cfg.To = entry.NewRecordField("NewNested")
				return cfg
			}(),
		},
		{
			"MoveFromNestedObj",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("nested", "nestedkey")
				cfg.To = entry.NewRecordField("unnestedkey")
				return cfg
			}(),
		},
		{
			"MoveToNestedObj",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("newnestedkey")
				cfg.To = entry.NewRecordField("nested", "newnestedkey")
				return cfg
			}(),
		},
		{
			"MoveDoubleNestedObj",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("nested", "nested2")
				cfg.To = entry.NewRecordField("nested2")
				return cfg
			}(),
		},
		{
			"MoveNestToResource",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("nested")
				cfg.To = entry.NewResourceField("NewNested")
				return cfg
			}(),
		},
		{
			"MoveNestToAttribute",
			false,
			func() *MoveOperatorConfig {
				cfg := defaultCfg()
				cfg.From = entry.NewRecordField("nested")
				cfg.To = entry.NewAttributeField("NewNested")
				return cfg
			}(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFromYaml, yamlErr := configFromFileViaYaml(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			cfgFromMapstructure, mapErr := configFromFileViaMapstructure(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			if tc.expectErr {
				require.Error(t, yamlErr)
				require.Error(t, mapErr)
			} else {
				require.NoError(t, yamlErr)
				require.Equal(t, tc.expect, cfgFromYaml)
				require.NoError(t, mapErr)
				require.Equal(t, tc.expect, cfgFromMapstructure)
			}
		})
	}
}

func configFromFileViaYaml(file string) (*MoveOperatorConfig, error) {
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

func configFromFileViaMapstructure(file string) (*MoveOperatorConfig, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}

	raw := map[string]interface{}{}

	if err := yaml.Unmarshal(bytes, raw); err != nil {
		return nil, fmt.Errorf("failed to read data from yaml: %s", err)
	}

	cfg := defaultCfg()
	dc := &mapstructure.DecoderConfig{Result: cfg, DecodeHook: helper.JSONUnmarshalerHook()}
	ms, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return nil, err
	}
	err = ms.Decode(raw)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultCfg() *MoveOperatorConfig {
	return NewMoveOperatorConfig("move")
}
