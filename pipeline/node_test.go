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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/input/generate"
	_ "github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/transformer/noop"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func TestNodeDOTID(t *testing.T) {
	operator := testutil.NewMockOperator("test")
	operator.On("Outputs").Return(nil)
	node := createOperatorNode(operator)
	require.Equal(t, operator.ID(), node.DOTID())
}

func TestCreateNodeID(t *testing.T) {
	nodeID := createNodeID("test_id")
	require.Equal(t, int64(5795108767401590291), nodeID)
}
