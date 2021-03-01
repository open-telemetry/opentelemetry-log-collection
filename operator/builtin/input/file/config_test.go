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

package file

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type testCase struct {
	name   string
	expect *InputConfig
}

func TestConfig(t *testing.T) {
	cases := []testCase{
		{
			"default",
			defaultCfg(),
		},
		{
			"id_custom",
			NewInputConfig("test_id"),
		},
		{
			"include_one",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				return cfg
			}(),
		},
		{
			"include_multi",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log", "two.log", "three.log")
				return cfg
			}(),
		},
		{
			"include_glob",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				return cfg
			}(),
		},
		{
			"include_inline",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "a.log", "b.log")
				return cfg
			}(),
		},
		{
			"exclude_one",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "one.log")
				return cfg
			}(),
		},
		{
			"exclude_multi",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "one.log", "two.log", "three.log")
				return cfg
			}(),
		},
		{
			"exclude_glob",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "not*.log")
				return cfg
			}(),
		},
		{
			"exclude_inline",
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "a.log", "b.log")
				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFromYaml, err := configFromFileViaYaml(t, path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			require.NoError(t, err)
			require.Equal(t, tc.expect, cfgFromYaml)

			// TODO cfgFromMapstructure, err := configFromFileViaYaml(t, path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			// TODO require.NoError(t, err)
			// TODO require.Equal(t, tc.expect, cfgFromMapstructure)
		})
	}
}

func configFromFileViaYaml(t *testing.T, file string) (*InputConfig, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}

	config := NewInputConfig("file_input")
	if err := yaml.UnmarshalStrict(bytes, config); err != nil {
		return nil, fmt.Errorf("failed to read config file as yaml: %s", err)
	}

	return config, nil
}

// TODO func configFromFileViaMapstructure(t *testing.T, file string) (*InputConfig, error) {}

func defaultCfg() *InputConfig {
	return NewInputConfig("file_input")
}
