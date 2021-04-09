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

package stdout

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

// Stdout is a global handle to standard output
var Stdout io.Writer = os.Stdout

func init() {
	operator.Register("stdout", func() operator.Builder { return NewStdoutConfig("") })
}

// NewStdoutConfig creates a new stdout config with default values
func NewStdoutConfig(operatorID string) *StdoutConfig {
	return &StdoutConfig{
		OutputConfig: helper.NewOutputConfig(operatorID, "stdout"),
	}
}

// StdoutConfig is the configuration of the Stdout operator
type StdoutConfig struct {
	helper.OutputConfig `yaml:",inline"`
}

// Build will build a stdout operator.
func (c StdoutConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	outputOperator, err := c.OutputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	op := &StdoutOperator{
		OutputOperator: outputOperator,
		encoder:        json.NewEncoder(Stdout),
	}
	return []operator.Operator{op}, nil
}

// StdoutOperator is an operator that logs entries using stdout.
type StdoutOperator struct {
	helper.OutputOperator
	encoder *json.Encoder
	mux     sync.Mutex
}

// Process will log entries received.
func (o *StdoutOperator) Process(ctx context.Context, entry *entry.Entry) error {
	o.mux.Lock()
	err := o.encoder.Encode(entry)
	if err != nil {
		o.mux.Unlock()
		o.Errorf("Failed to process entry: %s, $s", err, entry.Body)
		return err
	}
	o.mux.Unlock()
	return nil
}
