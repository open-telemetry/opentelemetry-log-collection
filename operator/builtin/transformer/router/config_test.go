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
package router

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper/operatortest"
)

func TestRouterGoldenConfig(t *testing.T) {
	cases := []operatortest.ConfigTestCase{
		{
			Name:   "default",
			Expect: defaultCfg(),
		},
		{
			Name: "routes_one",
			Expect: func() *RouterOperatorConfig {
				cfg := defaultCfg()
				newRoute := &RouterOperatorRouteConfig{
					Expression: `$.format == "json"`,
					OutputIDs:  []string{"my_json_parser"},
				}
				cfg.Routes = append(cfg.Routes, newRoute)
				return cfg
			}(),
		},
		{
			Name: "routes_multi",
			Expect: func() *RouterOperatorConfig {
				cfg := defaultCfg()
				newRoute := []*RouterOperatorRouteConfig{
					{
						Expression: `$.format == "json"`,
						OutputIDs:  []string{"my_json_parser"},
					},
					{
						Expression: `$.format == "json"2`,
						OutputIDs:  []string{"my_json_parser2"},
					},
					{
						Expression: `$.format == "json"3`,
						OutputIDs:  []string{"my_json_parser3"},
					},
				}
				cfg.Routes = newRoute
				return cfg
			}(),
		},
		{
			Name: "routes_attributes",
			Expect: func() *RouterOperatorConfig {
				cfg := defaultCfg()

				attVal := helper.NewAttributerConfig()
				attVal.Attributes = map[string]helper.ExprStringConfig{
					"key1": "val1",
				}

				cfg.Routes = []*RouterOperatorRouteConfig{
					{
						Expression:       `$.format == "json"`,
						OutputIDs:        []string{"my_json_parser"},
						AttributerConfig: attVal,
					},
				}
				return cfg
			}(),
		},
		{
			Name: "routes_default",
			Expect: func() *RouterOperatorConfig {
				cfg := defaultCfg()
				newRoute := &RouterOperatorRouteConfig{
					Expression: `$.format == "json"`,
					OutputIDs:  []string{"my_json_parser"},
				}
				cfg.Routes = append(cfg.Routes, newRoute)
				cfg.Default = append(cfg.Default, "catchall")
				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			operatortest.RunGoldenConfigTest(t, defaultCfg(), tc)
		})
	}
}

func defaultCfg() *RouterOperatorConfig {
	return NewRouterOperatorConfig("router")
}
