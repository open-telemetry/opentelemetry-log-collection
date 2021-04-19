package add

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
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper/operatortest"
)

type configTestCase struct {
	name   string
	expect *AddOperatorConfig
}

func TestGoldenConfig(t *testing.T) {
	cases := []configTestCase{
		{
			"add_value",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = "randomMessage"
				return cfg
			}(),
		},
		{
			"add_expr",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = `EXPR($.key + "_suffix")`
				return cfg
			}(),
		},
		{
			"add_nest",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = map[interface{}]interface{}{
					"nest": map[interface{}]interface{}{"key": "val"},
				}
				return cfg
			}(),
		},
		{
			"add_attribute",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewAttributeField("new")
				cfg.Value = "newVal"
				return cfg
			}(),
		},
		{
			"add_resource",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = "newVal"
				return cfg
			}(),
		},
		{
			"add_resource_expr",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("new")
				cfg.Value = `EXPR($.key + "_suffix")`
				return cfg
			}(),
		},
		{
			"add_array_to_body",
			func() *AddOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("new")
				cfg.Value = []interface{}{1, 2, 3, 4}
				return cfg
			}(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFromYaml, yamlErr := operatortest.ConfigFromFileViaYaml(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)), defaultCfg())
			cfgFromMapstructure, mapErr := operatortest.ConfigFromFileViaMapstructure(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)), defaultCfg())
			require.NoError(t, yamlErr)
			require.Equal(t, tc.expect, cfgFromYaml)
			require.NoError(t, mapErr)
			require.Equal(t, tc.expect, cfgFromMapstructure)
		})
	}
}

func defaultCfg() *AddOperatorConfig {
	return NewAddOperatorConfig("add")
}
