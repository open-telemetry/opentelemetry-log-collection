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

package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

type fileInputBenchmark struct {
	name   string
	paths  []string
	config func() *InputConfig
}

type benchFile struct {
	*os.File
	log func(int)
}

func simpleTextFile(file *os.File) *benchFile {
	line := stringWithLength(49) + "\n"
	return &benchFile{
		File: file,
		log:  func(_ int) { file.WriteString(line) },
	}
}

func BenchmarkFileInput(b *testing.B) {
	cases := []fileInputBenchmark{
		{
			name: "Single",
			paths: []string{
				"file0.log",
			},
			config: func() *InputConfig {
				cfg := NewInputConfig("test_id")
				cfg.Include = []string{
					"file0.log",
				}
				return cfg
			},
		},
		{
			name: "Glob",
			paths: []string{
				"file0.log",
				"file1.log",
				"file2.log",
				"file3.log",
			},
			config: func() *InputConfig {
				cfg := NewInputConfig("test_id")
				cfg.Include = []string{"file*.log"}
				return cfg
			},
		},
		{
			name: "MultiGlob",
			paths: []string{
				"file0.log",
				"file1.log",
				"log0.log",
				"log1.log",
			},
			config: func() *InputConfig {
				cfg := NewInputConfig("test_id")
				cfg.Include = []string{
					"file*.log",
					"log*.log",
				}
				return cfg
			},
		},
		{
			name: "MaxConcurrent",
			paths: []string{
				"file0.log",
				"file1.log",
				"file2.log",
				"file3.log",
			},
			config: func() *InputConfig {
				cfg := NewInputConfig("test_id")
				cfg.Include = []string{
					"file*.log",
				}
				cfg.MaxConcurrentFiles = 2
				return cfg
			},
		},
		{
			name: "FngrPrntLarge",
			paths: []string{
				"file0.log",
			},
			config: func() *InputConfig {
				cfg := NewInputConfig("test_id")
				cfg.Include = []string{
					"file*.log",
				}
				cfg.FingerprintSize = 10 * defaultFingerprintSize
				return cfg
			},
		},
		{
			name: "FngrPrntSmall",
			paths: []string{
				"file0.log",
			},
			config: func() *InputConfig {
				cfg := NewInputConfig("test_id")
				cfg.Include = []string{
					"file*.log",
				}
				cfg.FingerprintSize = defaultFingerprintSize / 10
				return cfg
			},
		},
	}

	for _, bench := range cases {
		b.Run(bench.name, func(b *testing.B) {
			rootDir, err := ioutil.TempDir("", "")
			require.NoError(b, err)

			files := []*benchFile{}
			for _, path := range bench.paths {
				file := openFile(b, filepath.Join(rootDir, path))
				files = append(files, simpleTextFile(file))
			}

			cfg := bench.config()
			cfg.OutputIDs = []string{"fake"}
			for i, inc := range cfg.Include {
				cfg.Include[i] = filepath.Join(rootDir, inc)
			}
			cfg.StartAt = "beginning"

			op, err := cfg.Build(testutil.NewBuildContext(b))
			require.NoError(b, err)

			fakeOutput := testutil.NewFakeOutput(b)
			err = op.SetOutputs([]operator.Operator{fakeOutput})
			require.NoError(b, err)

			// write half the lines before starting
			mid := b.N / 2
			for i := 0; i < mid; i++ {
				for _, file := range files {
					file.log(i)
				}
			}

			b.ResetTimer()
			err = op.Start(testutil.NewMockPersister("test"))
			defer op.Stop()
			require.NoError(b, err)

			// write the remainder of lines while running
			go func() {
				for i := mid; i < b.N; i++ {
					for _, file := range files {
						file.log(i)
					}
				}
			}()

			for i := 0; i < b.N*len(files); i++ {
				<-fakeOutput.Received
			}
		})
	}
}
