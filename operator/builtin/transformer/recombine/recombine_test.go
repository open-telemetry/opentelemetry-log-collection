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

package recombine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func TestRecombineOperator(t *testing.T) {
	t1 := time.Date(2020, time.April, 11, 21, 34, 01, 0, time.UTC)
	t2 := time.Date(2020, time.April, 11, 21, 34, 02, 0, time.UTC)

	entryWithBody := func(ts time.Time, body interface{}) *entry.Entry {
		e := entry.New()
		e.Timestamp = ts
		e.Body = body
		return e
	}

	entryWithBodyAttr := func(ts time.Time, body interface{}, Attr map[string]string) *entry.Entry {
		e := entryWithBody(ts, body)
		for k, v := range Attr {
			e.AddAttribute(k, v)
		}
		return e
	}

	cases := []struct {
		name           string
		config         *RecombineOperatorConfig
		input          []*entry.Entry
		expectedOutput []*entry.Entry
	}{
		{
			"NoEntriesFirst",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsFirstEntry = "true"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			nil,
			nil,
		},
		{
			"NoEntriesLast",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "true"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			nil,
			nil,
		},
		{
			"OneEntryFirst",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsFirstEntry = "true"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			[]*entry.Entry{entry.New()},
			nil,
		},
		{
			"OneEntryLast",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "true"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			[]*entry.Entry{entryWithBody(t1, "test")},
			[]*entry.Entry{entryWithBody(t1, "test")},
		},
		{
			"TwoEntriesLast",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "$body == 'test2'"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBody(t1, "test1"),
				entryWithBody(t2, "test2"),
			},
			[]*entry.Entry{entryWithBody(t1, "test1\ntest2")},
		},
		{
			"ThreeEntriesFirstNewest",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsFirstEntry = "$body == 'test1'"
				cfg.OutputIDs = []string{"fake"}
				cfg.OverwriteWith = "newest"
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBody(t1, "test1"),
				entryWithBody(t2, "test2"),
				entryWithBody(t2, "test1"),
			},
			[]*entry.Entry{
				entryWithBody(t2, "test1\ntest2"),
			},
		},
		{
			"CombineWithEmptyString",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.CombineWith = ""
				cfg.IsLastEntry = "$body == 'test2'"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBody(t1, "test1"),
				entryWithBody(t1, "test2"),
			},
			[]*entry.Entry{entryWithBody(t1, "test1test2")},
		},
		{
			"TestDefaultSourceIdentifier",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "$body == 'end'"
				cfg.OutputIDs = []string{"fake"}
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t1, "file2", map[string]string{"file.path": "file2"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file2"}),
			},
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1\nend", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t1, "file2\nend", map[string]string{"file.path": "file2"}),
			},
		},
		{
			"TestCustomSourceIdentifier",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "$body == 'end'"
				cfg.OutputIDs = []string{"fake"}
				cfg.SourceIdentifier = entry.NewAttributeField("custom_source")
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1", map[string]string{"custom_source": "file1"}),
				entryWithBodyAttr(t1, "file2", map[string]string{"custom_source": "file2"}),
				entryWithBodyAttr(t2, "end", map[string]string{"custom_source": "file1"}),
				entryWithBodyAttr(t2, "end", map[string]string{"custom_source": "file2"}),
			},
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1\nend", map[string]string{"custom_source": "file1"}),
				entryWithBodyAttr(t1, "file2\nend", map[string]string{"custom_source": "file2"}),
			},
		},
		{
			"TestMaxSources",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "$body == 'end'"
				cfg.OutputIDs = []string{"fake"}
				cfg.MaxSources = 1
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file1"}),
			},
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file1"}),
			},
		},
		{
			"TestMaxBatchSize",
			func() *RecombineOperatorConfig {
				cfg := NewRecombineOperatorConfig("")
				cfg.CombineField = entry.NewBodyField()
				cfg.IsLastEntry = "$body == 'end'"
				cfg.OutputIDs = []string{"fake"}
				cfg.MaxBatchSize = 2
				return cfg
			}(),
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1_event1", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t1, "file2_event1", map[string]string{"file.path": "file2"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t2, "file2_event2", map[string]string{"file.path": "file2"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file2"}),
			},
			[]*entry.Entry{
				entryWithBodyAttr(t1, "file1_event1\nend", map[string]string{"file.path": "file1"}),
				entryWithBodyAttr(t1, "file2_event1\nfile2_event2", map[string]string{"file.path": "file2"}),
				entryWithBodyAttr(t2, "end", map[string]string{"file.path": "file2"}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op, err := tc.config.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			recombine := op.(*RecombineOperator)

			fake := testutil.NewFakeOutput(t)
			err = recombine.SetOutputs([]operator.Operator{fake})
			require.NoError(t, err)

			for _, e := range tc.input {
				recombine.Process(context.Background(), e)
			}

			for _, expected := range tc.expectedOutput {
				fake.ExpectEntry(t, expected)
			}

			select {
			case e := <-fake.Received:
				require.FailNow(t, "Received unexpected entry: ", e)
			default:
			}
		})
	}

	t.Run("FlushesOnShutdown", func(t *testing.T) {
		cfg := NewRecombineOperatorConfig("")
		cfg.CombineField = entry.NewBodyField()
		cfg.IsFirstEntry = "false"
		cfg.OutputIDs = []string{"fake"}
		op, err := cfg.Build(testutil.NewBuildContext(t))
		require.NoError(t, err)
		recombine := op.(*RecombineOperator)

		fake := testutil.NewFakeOutput(t)
		err = recombine.SetOutputs([]operator.Operator{fake})
		require.NoError(t, err)

		// Send an entry that isn't the last in a multiline
		recombine.Process(context.Background(), entry.New())

		// Ensure that the entry isn't immediately sent
		select {
		case <-fake.Received:
			require.FailNow(t, "Received unexpected entry")
		case <-time.After(10 * time.Millisecond):
		}

		// Stop the operator
		recombine.Stop()

		// Ensure that the entries in the buffer are flushed
		select {
		case <-fake.Received:
		default:
			require.FailNow(t, "Entry was not flushed on shutdown")
		}
	})
}

func BenchmarkRecombine(b *testing.B) {
	cfg := NewRecombineOperatorConfig("")
	cfg.CombineField = entry.NewBodyField()
	cfg.IsFirstEntry = "false"
	cfg.OutputIDs = []string{"fake"}
	op, err := cfg.Build(testutil.NewBuildContext(b))
	require.NoError(b, err)
	recombine := op.(*RecombineOperator)

	fake := testutil.NewFakeOutput(b)
	require.NoError(b, recombine.SetOutputs([]operator.Operator{fake}))
	recombine.Start(nil)

	go func() {
		for {
			<-fake.Received
		}
	}()

	e := entry.New()
	e.Timestamp = time.Now()
	e.Body = "body"

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, recombine.Process(ctx, e))
		require.NoError(b, recombine.Process(ctx, e))
		require.NoError(b, recombine.Process(ctx, e))
		require.NoError(b, recombine.Process(ctx, e))
		require.NoError(b, recombine.Process(ctx, e))
		recombine.flushUncombined(ctx)
	}
}

func TestTimeout(t *testing.T) {
	t.Parallel()

	cfg := NewRecombineOperatorConfig("")
	cfg.CombineField = entry.NewBodyField()
	cfg.IsFirstEntry = "false"
	cfg.OutputIDs = []string{"fake"}
	cfg.ForceFlushTimeout = 100 * time.Millisecond
	op, err := cfg.Build(testutil.NewBuildContext(t))
	require.NoError(t, err)
	recombine := op.(*RecombineOperator)

	fake := testutil.NewFakeOutput(t)
	require.NoError(t, recombine.SetOutputs([]operator.Operator{fake}))

	e := entry.New()
	e.Timestamp = time.Now()
	e.Body = "body"

	ctx := context.Background()

	recombine.Start(nil)
	require.NoError(t, recombine.Process(ctx, e))
	select {
	case <-fake.Received:
		t.Logf("We shouldn't receive an entry before timeout")
		t.FailNow()
	case <-time.After(50 * time.Millisecond):
	}

	select {
	case <-fake.Received:
	case <-time.After(5 * time.Second):
		t.Logf("The entry should be flushed by now")
		t.FailNow()
	}

	require.NoError(t, recombine.Stop())
}
