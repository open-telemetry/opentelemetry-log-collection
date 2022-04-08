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
package regex

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper/operatortest"
)

func TestRegexParserGoldenConfig(t *testing.T) {
	cases := []operatortest.ConfigUnmarshalTest{
		{
			Name:   "default",
			Expect: defaultCfg(),
		},
		{
			Name: "cache",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				cfg.Cache.Size = 50
				return cfg
			}(),
		},
		{
			Name: "parse_from_simple",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				cfg.ParseFrom = entry.NewBodyField("from")
				return cfg
			}(),
		},
		{
			Name: "parse_to_simple",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				cfg.ParseTo = entry.NewBodyField("log")
				return cfg
			}(),
		},
		{
			Name: "on_error_drop",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				cfg.OnError = "drop"
				return cfg
			}(),
		},
		{
			Name: "timestamp",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				parseField := entry.NewBodyField("timestamp_field")
				newTime := helper.TimeParser{
					LayoutType: "strptime",
					Layout:     "%Y-%m-%d",
					ParseFrom:  &parseField,
				}
				cfg.TimeParser = &newTime
				return cfg
			}(),
		},
		{
			Name: "severity",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				parseField := entry.NewBodyField("severity_field")
				severityParser := helper.NewSeverityParserConfig()
				severityParser.ParseFrom = &parseField
				mapping := map[interface{}]interface{}{
					"critical": "5xx",
					"error":    "4xx",
					"info":     "3xx",
					"debug":    "2xx",
				}
				severityParser.Mapping = mapping
				cfg.SeverityParserConfig = &severityParser
				return cfg
			}(),
		},
		{
			Name: "preserve_to",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				preserve := entry.NewBodyField("aField")
				cfg.PreserveTo = &preserve
				return cfg
			}(),
		},
		{
			Name: "regex",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				cfg.Regex = "^Host=(?P<host>[^,]+), Type=(?P<type>.*)$"
				return cfg
			}(),
		},
		{
			Name: "scope_name",
			Expect: func() *RegexParserConfig {
				cfg := defaultCfg()
				cfg.Regex = "^Host=(?P<host>[^,]+), Logger=(?P<logger_name_field>.*)$"
				parseField := entry.NewBodyField("logger_name_field")
				loggerNameParser := helper.NewScopeNameParser()
				loggerNameParser.ParseFrom = parseField
				cfg.ScopeNameParser = &loggerNameParser
				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Run(t, defaultCfg())
		})
	}
}

func defaultCfg() *RegexParserConfig {
	return NewRegexParserConfig("regex_parser")
}
