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
