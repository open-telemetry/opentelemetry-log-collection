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
package csv

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper/operatortest"
)

func TestJSONParserConfig(t *testing.T) {
	cases := []operatortest.ConfigUnmarshalTest{
		{
			Name: "basic",
			Expect: func() *CSVParserConfig {
				p := defaultCfg()
				p.Header = "id,severity,message"
				p.ParseFrom = entry.NewBodyField("message")
				return p
			}(),
		},
		{
			Name: "delimiter",
			Expect: func() *CSVParserConfig {
				p := defaultCfg()
				p.Header = "id,severity,message"
				p.ParseFrom = entry.NewBodyField("message")
				p.FieldDelimiter = "\t"
				return p
			}(),
		},
		{
			Name: "timestamp",
			Expect: func() *CSVParserConfig {
				p := defaultCfg()
				p.Header = "timestamp_field,severity,message"
				newTime := helper.NewTimeParser()
				p.TimeParser = &newTime
				parseFrom := entry.NewBodyField("timestamp_field")
				p.TimeParser.ParseFrom = &parseFrom
				p.TimeParser.LayoutType = "strptime"
				p.TimeParser.Layout = "%Y-%m-%d"
				return p
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Run(t, defaultCfg())
		})
	}
}

func defaultCfg() *CSVParserConfig {
	return NewCSVParserConfig("json_parser")
}
