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
package remove

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper/operatortest"
)

// test unmarshalling of values into config struct
func TestGoldenConfig(t *testing.T) {
	cases := []operatortest.ConfigTestCase{
		{
			Name: "remove_body",
			Expect: func() *RemoveOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewBodyField("nested")
				return cfg
			}(),
		},
		{
			Name: "remove_single_attribute",
			Expect: func() *RemoveOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewAttributeField("key")
				return cfg
			}(),
		},
		{
			Name: "remove_single_resource",
			Expect: func() *RemoveOperatorConfig {
				cfg := defaultCfg()
				cfg.Field = entry.NewResourceField("key")
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

func defaultCfg() *RemoveOperatorConfig {
	return NewRemoveOperatorConfig("move")
}
