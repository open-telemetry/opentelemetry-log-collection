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
	"context"
	"os"
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRouterOperator(t *testing.T) {
	os.Setenv("TEST_ROUTER_PLUGIN_ENV", "foo")
	defer os.Unsetenv("TEST_ROUTER_PLUGIN_ENV")

	basicConfig := func() *RouterOperatorConfig {
		return &RouterOperatorConfig{
			BasicConfig: helper.BasicConfig{
				OperatorID:   "test_operator_id",
				OperatorType: "router",
			},
		}
	}

	cases := []struct {
		name               string
		input              *entry.Entry
		routes             []*RouterOperatorRouteConfig
		defaultOutput      helper.OutputIDs
		expectedCounts     map[string]int
		expectedAttributes map[string]string
	}{
		{
			"DefaultRoute",
			entry.New(),
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					"true",
					[]string{"output1"},
				},
			},
			nil,
			map[string]int{"output1": 1},
			nil,
		},
		{
			"NoMatch",
			entry.New(),
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					`false`,
					[]string{"output1"},
				},
			},
			nil,
			map[string]int{},
			nil,
		},
		{
			"SimpleMatch",
			&entry.Entry{
				Record: map[string]interface{}{
					"message": "test_message",
				},
			},
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					`$.message == "non_match"`,
					[]string{"output1"},
				},
				{
					helper.NewAttributerConfig(),
					`$.message == "test_message"`,
					[]string{"output2"},
				},
			},
			nil,
			map[string]int{"output2": 1},
			nil,
		},
		{
			"MatchWithAttribute",
			&entry.Entry{
				Record: map[string]interface{}{
					"message": "test_message",
				},
			},
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					`$.message == "non_match"`,
					[]string{"output1"},
				},
				{
					helper.AttributerConfig{
						Attributes: map[string]helper.ExprStringConfig{
							"label-key": "label-value",
						},
					},
					`$.message == "test_message"`,
					[]string{"output2"},
				},
			},
			nil,
			map[string]int{"output2": 1},
			map[string]string{
				"label-key": "label-value",
			},
		},
		{
			"MatchEnv",
			&entry.Entry{
				Record: map[string]interface{}{
					"message": "test_message",
				},
			},
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					`env("TEST_ROUTER_PLUGIN_ENV") == "foo"`,
					[]string{"output1"},
				},
				{
					helper.NewAttributerConfig(),
					`true`,
					[]string{"output2"},
				},
			},
			nil,
			map[string]int{"output1": 1},
			nil,
		},
		{
			"UseDefault",
			&entry.Entry{
				Record: map[string]interface{}{
					"message": "test_message",
				},
			},
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					`false`,
					[]string{"output1"},
				},
			},
			[]string{"output2"},
			map[string]int{"output2": 1},
			nil,
		},
		{
			"MatchBeforeDefault",
			&entry.Entry{
				Record: map[string]interface{}{
					"message": "test_message",
				},
			},
			[]*RouterOperatorRouteConfig{
				{
					helper.NewAttributerConfig(),
					`true`,
					[]string{"output1"},
				},
			},
			[]string{"output2"},
			map[string]int{"output1": 1},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := basicConfig()
			cfg.Routes = tc.routes
			cfg.Default = tc.defaultOutput

			buildContext := testutil.NewBuildContext(t)
			ops, err := cfg.Build(buildContext)
			require.NoError(t, err)
			op := ops[0]

			results := map[string]int{}
			var attributes map[string]string

			mock1 := testutil.NewMockOperator("$.output1")
			mock1.On("Process", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
				results["output1"]++
				if entry, ok := args[1].(*entry.Entry); ok {
					attributes = entry.Attributes
				}
			})

			mock2 := testutil.NewMockOperator("$.output2")
			mock2.On("Process", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
				results["output2"]++
				if entry, ok := args[1].(*entry.Entry); ok {
					attributes = entry.Attributes
				}
			})

			routerOperator := op.(*RouterOperator)
			err = routerOperator.SetOutputs([]operator.Operator{mock1, mock2})
			require.NoError(t, err)

			err = routerOperator.Process(context.Background(), tc.input)
			require.NoError(t, err)

			require.Equal(t, tc.expectedCounts, results)
			require.Equal(t, tc.expectedAttributes, attributes)
		})
	}
}
