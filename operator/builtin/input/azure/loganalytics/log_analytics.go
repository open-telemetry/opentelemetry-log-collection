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

package loganalytics

import (
	"context"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/input/azure"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

const operatorName = "azure_log_analytics_input"

func init() {
	operator.Register(operatorName, func() operator.Builder { return NewLogAnalyticsConfig("") })
}

// NewLogAnalyticsConfig creates a new Azure Log Analytics input config with default values
func NewLogAnalyticsConfig(operatorID string) *LogAnalyticsInputConfig {
	return &LogAnalyticsInputConfig{
		InputConfig: helper.NewInputConfig(operatorID, operatorName),
		AzureConfig: azure.AzureConfig{
			PrefetchCount: 1000,
			StartAt:       "end",
		},
	}
}

// LogAnalyticsInputConfig is the configuration of a Azure Log Analytics input operator.
type LogAnalyticsInputConfig struct {
	helper.InputConfig `yaml:",inline"`
	azure.AzureConfig  `yaml:",inline"`
}

// Build will build a Azure Log Analytics input operator.
func (c *LogAnalyticsInputConfig) Build(buildContext operator.BuildContext) ([]operator.Operator, error) {
	if err := c.AzureConfig.Build(buildContext, c.InputConfig); err != nil {
		return nil, err
	}

	logAnalyticsInput := &LogAnalyticsInput{
		EventHub: azure.EventHub{
			AzureConfig: c.AzureConfig,
		},
		json: jsoniter.ConfigFastest,
	}
	return []operator.Operator{logAnalyticsInput}, nil
}

// LogAnalyticsInput is an operator that reads Azure Log Analytics input from Azure Event Hub.
type LogAnalyticsInput struct {
	azure.EventHub
	json jsoniter.API
}

// Start will start generating log entries.
func (l *LogAnalyticsInput) Start(persister operator.Persister) error {
	l.Handler = l.handleBatchedEvents
	l.Persist = &azure.Persister{DB: persister}
	return l.StartConsumers(context.Background())
}

// Stop will stop generating logs.
func (l *LogAnalyticsInput) Stop() error {
	return l.StopConsumers()
}
