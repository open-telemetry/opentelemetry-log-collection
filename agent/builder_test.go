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

package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/transformer/noop"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func TestBuildAgentSuccess(t *testing.T) {
	mockCfg := Config{}
	mockLogger := zap.NewNop().Sugar()
	mockPluginDir := "/some/path/plugins"

	agent, err := NewBuilder(mockLogger).
		WithConfig(&mockCfg).
		WithPluginDir(mockPluginDir).
		Build()
	require.NoError(t, err)
	require.Equal(t, mockLogger, agent.SugaredLogger)
}

func TestBuildAgentDefaultOperator(t *testing.T) {
	mockCfg := Config{
		[]operator.Config{
			operator.Config{
				Builder: noop.NewNoopOperatorConfig("noop"),
			},
			operator.Config{
				Builder: noop.NewNoopOperatorConfig("noop1"),
			},
		},
	}
	mockLogger := zap.NewNop().Sugar()
	mockPluginDir := "/some/path/plugins"
	mockOutput := testutil.NewFakeOutput(t)

	agent, err := NewBuilder(mockLogger).
		WithConfig(&mockCfg).
		WithPluginDir(mockPluginDir).
		WithDefaultOutput(mockOutput).
		Build()
	require.NoError(t, err)
	require.Equal(t, mockLogger, agent.SugaredLogger)

	ops := agent.pipeline.Operators()
	require.Equal(t, 3, len(ops))

	exists := make(map[string]bool)

	for _, op := range ops {
		if op.ID() == "$.noop" {
			require.Equal(t, 1, len(ops[0].GetOutputIDs()))
			require.Equal(t, "$.noop1", ops[0].GetOutputIDs()[0])
			exists["$.noop"] = true
		} else if op.ID() == "$.noop1" {
			require.Equal(t, 1, len(ops[1].GetOutputIDs()))
			require.Equal(t, "$.fake", ops[1].GetOutputIDs()[0])
			exists["$.noop1"] = true
		} else if op.ID() == "$.fake" {
			require.Equal(t, 0, len(ops[2].GetOutputIDs()))
			exists["$.fake"] = true
		}
	}
	require.Equal(t, true, exists["$.noop"])
	require.Equal(t, true, exists["$.noop1"])
	require.Equal(t, true, exists["$.fake"])
}

func TestBuildAgentFailureOnPluginRegistry(t *testing.T) {
	mockCfg := Config{}
	mockLogger := zap.NewNop().Sugar()
	mockPluginDir := "[]"
	mockOutput := testutil.NewFakeOutput(t)

	_, err := NewBuilder(mockLogger).
		WithConfig(&mockCfg).
		WithPluginDir(mockPluginDir).
		WithDefaultOutput(mockOutput).
		Build()
	require.NoError(t, err)
}

func TestBuildAgentFailureNoConfigOrGlobs(t *testing.T) {
	mockLogger := zap.NewNop().Sugar()
	mockPluginDir := "/some/plugin/path"
	mockOutput := testutil.NewFakeOutput(t)

	agent, err := NewBuilder(mockLogger).
		WithPluginDir(mockPluginDir).
		WithDefaultOutput(mockOutput).
		Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be built without")
	require.Nil(t, agent)
}

func TestBuildAgentFailureWithBothConfigAndGlobs(t *testing.T) {
	mockCfg := Config{}
	mockLogger := zap.NewNop().Sugar()
	mockPluginDir := "/some/plugin/path"
	mockOutput := testutil.NewFakeOutput(t)

	agent, err := NewBuilder(mockLogger).
		WithConfig(&mockCfg).
		WithConfigFiles([]string{"test"}).
		WithPluginDir(mockPluginDir).
		WithDefaultOutput(mockOutput).
		Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not both")
	require.Nil(t, agent)
}

func TestBuildAgentFailureNonexistGlobs(t *testing.T) {
	mockLogger := zap.NewNop().Sugar()
	mockPluginDir := "/some/plugin/path"
	mockOutput := testutil.NewFakeOutput(t)

	agent, err := NewBuilder(mockLogger).
		WithConfigFiles([]string{"/tmp/nonexist"}).
		WithPluginDir(mockPluginDir).
		WithDefaultOutput(mockOutput).
		Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "read configs from globs")
	require.Nil(t, agent)
}
