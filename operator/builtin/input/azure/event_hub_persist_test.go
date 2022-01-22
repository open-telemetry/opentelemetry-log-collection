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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-event-hubs-go/v3/persist"
	"github.com/stretchr/testify/require"
)

type mockPersisterDB struct {
	data map[string][]byte
	wg   sync.WaitGroup
}

func (p *mockPersisterDB) Get(ctx context.Context, key string) ([]byte, error) {
	p.wg.Add(1)
	defer p.wg.Done()

	data, ok := p.data[key]
	if !ok {
		return nil, fmt.Errorf("failed to lookup key %s", key)
	}
	return data, nil
}

func (p *mockPersisterDB) Set(ctx context.Context, key string, data []byte) error {
	p.wg.Add(1)
	defer p.wg.Done()

	p.data[key] = data
	return nil
}

func (p *mockPersisterDB) Delete(ctx context.Context, key string) error {
	return nil
}

func TestReadWrite(t *testing.T) {
	mockDB := mockPersisterDB{}
	mockDB.data = make(map[string][]byte)

	persiter := Persister{
		DB: &mockDB,
	}

	inputCP := persist.Checkpoint{
		Offset:         "abc",
		SequenceNumber: 10,
		EnqueueTime:    time.Now().Local(),
	}

	err := persiter.Write(
		"namespace-otel",
		"name-otel",
		"group-otel",
		"partition-otel",
		inputCP,
	)
	require.NoError(t, err)

	outputCP, err := persiter.Read(
		"namespace-otel",
		"name-otel",
		"group-otel",
		"partition-otel",
	)
	require.NoError(t, err)

	require.Equal(t, inputCP, outputCP)
}

func TestPersistenceKey(t *testing.T) {
	type TestKey struct {
		namespace     string
		name          string
		consumerGroup string
		partitionID   string
	}

	cases := []struct {
		name     string
		input    TestKey
		expected string
	}{
		{
			"basic",
			TestKey{
				namespace:     "stanza",
				name:          "devel",
				consumerGroup: "$Default",
				partitionID:   "0",
			},
			"stanza-devel-$Default-0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := Persister{}
			out := p.persistenceKey(tc.input.namespace, tc.input.name, tc.input.consumerGroup, tc.input.partitionID)
			require.Equal(t, tc.expected, out)
		})
	}
}
