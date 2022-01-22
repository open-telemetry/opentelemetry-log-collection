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

package azure

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func TestBuild(t *testing.T) {
	cases := []struct {
		name        string
		inputConfig AzureConfig
		h           helper.InputConfig
		expectErr   bool
	}{
		{
			"valid",
			AzureConfig{
				Namespace:        "namespace-otel",
				Name:             "name-otel",
				Group:            "group-otel",
				ConnectionString: "connection-string-otel",
				StartAt:          "end",
				PrefetchCount:    1000,
			},
			helper.NewInputConfig("test", "test"),
			false,
		},
		{
			"start-at-beginning",
			AzureConfig{
				Namespace:        "namespace-otel",
				Name:             "name-otel",
				Group:            "group-otel",
				ConnectionString: "connection-string-otel",
				StartAt:          "beginning",
				PrefetchCount:    1000,
			},
			helper.NewInputConfig("test", "test"),
			false,
		},
		{
			"invalid-start-at",
			AzureConfig{
				Namespace:        "namespace-otel",
				Name:             "name-otel",
				Group:            "group-otel",
				ConnectionString: "connection-string-otel",
				StartAt:          "bad",
				PrefetchCount:    1000,
			},
			helper.NewInputConfig("test", "test"),
			true,
		},
		{
			"invalid-input-config",
			AzureConfig{
				Namespace:        "namespace-otel",
				Name:             "name-otel",
				Group:            "group-otel",
				ConnectionString: "connection-string-otel",
				StartAt:          "end",
				PrefetchCount:    1000,
			},
			helper.InputConfig{},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.inputConfig.Build(testutil.NewBuildContext(t), tc.h)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name      string
		input     AzureConfig
		expectErr bool
	}{
		{
			"missing-namespace",
			AzureConfig{
				Namespace:        "",
				Name:             "john",
				Group:            "devel",
				ConnectionString: "some connection string",
				StartAt:          "end",
				PrefetchCount:    10,
			},
			true,
		},
		{
			"missing-name",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "",
				Group:            "devel",
				ConnectionString: "some connection string",
				StartAt:          "end",
				PrefetchCount:    10,
			},
			true,
		},
		{
			"missing-group",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "dev",
				Group:            "",
				ConnectionString: "some connection string",
				StartAt:          "end",
				PrefetchCount:    10,
			},
			true,
		},
		{
			"missing-connection-string",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "dev",
				Group:            "dev",
				ConnectionString: "",
				StartAt:          "end",
				PrefetchCount:    10,
			},
			true,
		},
		{
			"invalid-prefetch-count",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "dev",
				Group:            "dev",
				ConnectionString: "some string",
				StartAt:          "end",
				PrefetchCount:    0,
			},
			true,
		},
		{
			"invalid-start-at",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "dev",
				Group:            "dev",
				ConnectionString: "some string",
				StartAt:          "bad",
				PrefetchCount:    10,
			},
			true,
		},
		{
			"valid-start-at-end",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "dev",
				Group:            "dev",
				ConnectionString: "some string",
				StartAt:          "end",
				PrefetchCount:    10,
			},
			false,
		},
		{
			"valid-start-at-beginning",
			AzureConfig{
				Namespace:        "namespace",
				Name:             "dev",
				Group:            "dev",
				ConnectionString: "some string",
				PrefetchCount:    10,
				StartAt:          "beginning",
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
