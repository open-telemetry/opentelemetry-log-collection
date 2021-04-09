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

package operator

//go:generate mockery --name=^(Operator)$ --output=../testutil --outpkg=testutil --case=snake

import (
	"context"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
)

// Operator is a log monitoring component.
type Operator interface {
	// ID returns the id of the operator.
	ID() string
	// Type returns the type of the operator.
	Type() string

	// Start will start the operator.
	Start(Persister) error
	// Stop will stop the operator.
	Stop() error

	// CanOutput indicates if the operator will output entries to other operators.
	CanOutput() bool
	// Outputs returns the list of connected outputs.
	Outputs() []Operator
	// SetOutputs will set the connected outputs.
	SetOutputs([]Operator) error

	// CanProcess indicates if the operator will process entries from other operators.
	CanProcess() bool
	// Process will process an entry from an operator.
	Process(context.Context, *entry.Entry) error
	// Logger returns the operator's logger
	Logger() *zap.SugaredLogger
}
