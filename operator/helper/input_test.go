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

package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func TestInputConfigMissingBase(t *testing.T) {
	config := InputConfig{
		WriteTo: entry.Field{},
		WriterConfig: WriterConfig{
			OutputIDs: []string{"test-output"},
		},
	}
	context := testutil.NewBuildContext(t)
	_, err := config.Build(context)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required `type` field.")
}

func TestInputConfigMissingOutput(t *testing.T) {
	config := InputConfig{
		WriterConfig: WriterConfig{
			BasicConfig: BasicConfig{
				OperatorID:   "test-id",
				OperatorType: "test-type",
			},
		},
		WriteTo: entry.Field{},
	}
	context := testutil.NewBuildContext(t)
	_, err := config.Build(context)
	require.NoError(t, err)
}

func TestInputConfigValid(t *testing.T) {
	config := InputConfig{
		WriteTo: entry.Field{},
		WriterConfig: WriterConfig{
			BasicConfig: BasicConfig{
				OperatorID:   "test-id",
				OperatorType: "test-type",
			},
			OutputIDs: []string{"test-output"},
		},
	}
	context := testutil.NewBuildContext(t)
	_, err := config.Build(context)
	require.NoError(t, err)
}

func TestInputOperatorCanProcess(t *testing.T) {
	buildContext := testutil.NewBuildContext(t)
	input := InputOperator{
		WriterOperator: WriterOperator{
			BasicOperator: BasicOperator{
				OperatorID:    "test-id",
				OperatorType:  "test-type",
				SugaredLogger: buildContext.Logger.SugaredLogger,
			},
		},
	}
	require.False(t, input.CanProcess())
}

func TestInputOperatorProcess(t *testing.T) {
	buildContext := testutil.NewBuildContext(t)
	input := InputOperator{
		WriterOperator: WriterOperator{
			BasicOperator: BasicOperator{
				OperatorID:    "test-id",
				OperatorType:  "test-type",
				SugaredLogger: buildContext.Logger.SugaredLogger,
			},
		},
	}
	entry := entry.New()
	ctx := context.Background()
	err := input.Process(ctx, entry)
	require.Error(t, err)
	require.Equal(t, err.Error(), "Operator can not process logs.")
}

func TestInputOperatorNewEntry(t *testing.T) {
	buildContext := testutil.NewBuildContext(t)
	writeTo := entry.NewBodyField("test-field")

	labelExpr, err := ExprStringConfig("test").Build()
	require.NoError(t, err)

	resourceExpr, err := ExprStringConfig("resource").Build()
	require.NoError(t, err)

	input := InputOperator{
		Attributer: Attributer{
			attributes: map[string]*ExprString{
				"test-label": labelExpr,
			},
		},
		Identifier: Identifier{
			resource: map[string]*ExprString{
				"resource-key": resourceExpr,
			},
		},
		WriterOperator: WriterOperator{
			BasicOperator: BasicOperator{
				OperatorID:    "test-id",
				OperatorType:  "test-type",
				SugaredLogger: buildContext.Logger.SugaredLogger,
			},
		},
		WriteTo: writeTo,
	}

	entry, err := input.NewEntry("test")
	require.NoError(t, err)

	value, exists := entry.Get(writeTo)
	require.True(t, exists)
	require.Equal(t, "test", value)

	labelValue, exists := entry.Attributes["test-label"]
	require.True(t, exists)
	require.Equal(t, "test", labelValue)

	resourceValue, exists := entry.Resource["resource-key"]
	require.True(t, exists)
	require.Equal(t, "resource", resourceValue)
}
