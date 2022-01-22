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
package eventhub

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper/operatortest"
)

func TestKVParserConfig(t *testing.T) {
	cases := []operatortest.ConfigUnmarshalTest{
		{
			Name:   "default",
			Expect: defaultCfg(),
		},
		{
			Name: "connectionstring",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.ConnectionString = "connection-string-otel"
				return cfg
			}(),
		},
		{
			Name: "group",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.Group = "group-otel"
				return cfg
			}(),
		},
		{
			Name: "name",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.Name = "name-otel"
				return cfg
			}(),
		},
		{
			Name: "namespace",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.Namespace = "namespace-otel"
				return cfg
			}(),
		},
		{
			Name: "prefetch_count_10",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.PrefetchCount = 10
				return cfg
			}(),
		},
		{
			Name: "start_at_beginning",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.StartAt = "beginning"
				return cfg
			}(),
		},
		{
			Name: "start_at_end",
			Expect: func() *EventHubInputConfig {
				cfg := defaultCfg()
				cfg.StartAt = "end"
				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Run(t, defaultCfg())
		})
	}
}

func defaultCfg() *EventHubInputConfig {
	return NewEventHubConfig("azure_event_hub_input")
}
