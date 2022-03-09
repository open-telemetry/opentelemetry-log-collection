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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

var testHeader = "name,sev,msg"

func newTestParser(t *testing.T) *CSVParser {
	cfg := NewCSVParserConfig("test")
	cfg.Header = testHeader
	op, err := cfg.Build(testutil.Logger(t))
	require.NoError(t, err)
	return op.(*CSVParser)
}

func TestCSVParserBuildFailure(t *testing.T) {
	cfg := NewCSVParserConfig("test")
	cfg.OnError = "invalid_on_error"
	_, err := cfg.Build(testutil.Logger(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid `on_error` field")
}

func TestCSVParserBuildFailureInvalidDelimiter(t *testing.T) {
	cfg := NewCSVParserConfig("test")
	cfg.Header = testHeader
	cfg.FieldDelimiter = ";;"
	_, err := cfg.Build(testutil.Logger(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid 'delimiter': ';;'")
}

func TestCSVParserBuildFailureBadHeaderConfig(t *testing.T) {
	cfg := NewCSVParserConfig("test")
	cfg.Header = "testheader"
	cfg.HeaderAttribute = "testheader"
	_, err := cfg.Build(testutil.Logger(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one header parameter can be set: 'header' or 'header_attribute'")
}

func TestCSVParserByteFailure(t *testing.T) {
	parser := newTestParser(t)
	_, err := parser.parse([]byte("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of fields")
}

func TestCSVParserStringFailure(t *testing.T) {
	parser := newTestParser(t)
	_, err := parser.parse("invalid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of fields")
}

func TestCSVParserInvalidType(t *testing.T) {
	parser := newTestParser(t)
	_, err := parser.parse([]int{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "type '[]int' cannot be parsed as csv")
}

func TestParserCSV(t *testing.T) {
	cases := []struct {
		name             string
		configure        func(*CSVParserConfig)
		inputEntry       []entry.Entry
		outputBody       []interface{}
		expectBuildErr   bool
		expectProcessErr bool
	}{
		{
			"basic",
			func(p *CSVParserConfig) {
				p.Header = testHeader
			},
			[]entry.Entry{
				{
					Body: "stanza,INFO,started agent",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name": "stanza",
					"sev":  "INFO",
					"msg":  "started agent",
				},
			},
			false,
			false,
		},
		{
			"basic-multiple-static-bodies",
			func(p *CSVParserConfig) {
				p.Header = testHeader
			},
			[]entry.Entry{
				{
					Body: "stanza,INFO,started agent",
				},
				{
					Body: "stanza,ERROR,agent killed",
				},
				{
					Body: "kernel,TRACE,oom",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name": "stanza",
					"sev":  "INFO",
					"msg":  "started agent",
				},
				map[string]interface{}{
					"name": "stanza",
					"sev":  "ERROR",
					"msg":  "agent killed",
				},
				map[string]interface{}{
					"name": "kernel",
					"sev":  "TRACE",
					"msg":  "oom",
				},
			},
			false,
			false,
		},
		{
			"advanced",
			func(p *CSVParserConfig) {
				p.Header = "name;address;age;phone;position"
				p.FieldDelimiter = ";"
			},
			[]entry.Entry{
				{
					Body: "stanza;Evergreen;1;555-5555;agent",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":     "stanza",
					"address":  "Evergreen",
					"age":      "1",
					"phone":    "555-5555",
					"position": "agent",
				},
			},
			false,
			false,
		},
		{
			"dynamic-fields",
			func(p *CSVParserConfig) {
				p.HeaderAttribute = "Fields"
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Attributes: map[string]interface{}{
						"Fields": "name,age,height,number",
					},
					Body: "stanza dev,1,400,555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza dev",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			false,
			false,
		},
		{
			"dynamic-fields-multiple-entries",
			func(p *CSVParserConfig) {
				p.HeaderAttribute = "Fields"
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Attributes: map[string]interface{}{
						"Fields": "name,age,height,number",
					},
					Body: "stanza dev,1,400,555-555-5555",
				},
				{
					Attributes: map[string]interface{}{
						"Fields": "x,y",
					},
					Body: "000100,2",
				},
				{
					Attributes: map[string]interface{}{
						"Fields": "a,b,c,d,e,f",
					},
					Body: "1,2,3,4,5,6",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza dev",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
				map[string]interface{}{
					"x": "000100",
					"y": "2",
				},
				map[string]interface{}{
					"a": "1",
					"b": "2",
					"c": "3",
					"d": "4",
					"e": "5",
					"f": "6",
				},
			},
			false,
			false,
		},
		{
			"dynamic-fields-tab",
			func(p *CSVParserConfig) {
				p.HeaderAttribute = "columns"
				p.FieldDelimiter = "\t"
			},
			[]entry.Entry{
				{
					Attributes: map[string]interface{}{
						"columns": "name	age	height	number",
					},
					Body: "stanza dev	1	400	555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza dev",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			false,
			false,
		},
		{
			"dynamic-fields-label-missing",
			func(p *CSVParserConfig) {
				p.HeaderAttribute = "Fields"
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Body: "stanza dev,1,400,555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza dev",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			false,
			true,
		},
		{
			"missing-header-field",
			func(p *CSVParserConfig) {
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Body: "stanza,1,400,555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			true,
			false,
		},
		{
			"mariadb-audit-log",
			func(p *CSVParserConfig) {
				p.Header = "timestamp,serverhost,username,host,connectionid,queryid,operation,database,object,retcode"
			},
			[]entry.Entry{
				{
					Body: "20210316 17:08:01,oiq-int-mysql,load,oiq-int-mysql.bluemedora.localnet,5,0,DISCONNECT,,,0",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"timestamp":    "20210316 17:08:01",
					"serverhost":   "oiq-int-mysql",
					"username":     "load",
					"host":         "oiq-int-mysql.bluemedora.localnet",
					"connectionid": "5",
					"queryid":      "0",
					"operation":    "DISCONNECT",
					"database":     "",
					"object":       "",
					"retcode":      "0",
				},
			},
			false,
			false,
		},
		{
			"empty field",
			func(p *CSVParserConfig) {
				p.Header = "name,address,age,phone,position"
			},
			[]entry.Entry{
				{
					Body: "stanza,Evergreen,,555-5555,agent",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":     "stanza",
					"address":  "Evergreen",
					"age":      "",
					"phone":    "555-5555",
					"position": "agent",
				},
			},
			false,
			false,
		},
		{
			"tab delimiter",
			func(p *CSVParserConfig) {
				p.Header = "name	address	age	phone	position"
				p.FieldDelimiter = "\t"
			},
			[]entry.Entry{
				{
					Body: "stanza	Evergreen	1	555-5555	agent",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":     "stanza",
					"address":  "Evergreen",
					"age":      "1",
					"phone":    "555-5555",
					"position": "agent",
				},
			},
			false,
			false,
		},
		{
			"comma in quotes",
			func(p *CSVParserConfig) {
				p.Header = "name,address,age,phone,position"
			},
			[]entry.Entry{
				{
					Body: "stanza,\"Evergreen,49508\",1,555-5555,agent",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":     "stanza",
					"address":  "Evergreen,49508",
					"age":      "1",
					"phone":    "555-5555",
					"position": "agent",
				},
			},
			false,
			false,
		},
		{
			"quotes in quotes",
			func(p *CSVParserConfig) {
				p.Header = "name,address,age,phone,position"
			},
			[]entry.Entry{
				{
					Body: "\"bob \"\"the man\"\"\",Evergreen,1,555-5555,agent",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":     "bob \"the man\"",
					"address":  "Evergreen",
					"age":      "1",
					"phone":    "555-5555",
					"position": "agent",
				},
			},
			false,
			false,
		},
		{
			"missing-header-delimiter-in-header",
			func(p *CSVParserConfig) {
				p.Header = "name:age:height:number"
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Body: "stanza,1,400,555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			true,
			false,
		},
		{
			"invalid-delimiter",
			func(p *CSVParserConfig) {
				// expect []rune of length 1
				p.Header = "name,,age,,height,,number"
				p.FieldDelimiter = ",,"
			},
			[]entry.Entry{
				{
					Body: "stanza,1,400,555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			true,
			false,
		},
		{
			"parse-failure-num-fields-mismatch",
			func(p *CSVParserConfig) {
				p.Header = "name,age,height,number"
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Body: "1,400,555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			false,
			true,
		},
		{
			"parse-failure-wrong-field-delimiter",
			func(p *CSVParserConfig) {
				p.Header = "name,age,height,number"
				p.FieldDelimiter = ","
			},
			[]entry.Entry{
				{
					Body: "stanza:1:400:555-555-5555",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza",
					"age":    "1",
					"height": "400",
					"number": "555-555-5555",
				},
			},
			false,
			true,
		},
		{
			"parse-with-lazy-quotes",
			func(p *CSVParserConfig) {
				p.Header = "name,age,height,number"
				p.FieldDelimiter = ","
				p.LazyQuotes = true
			},
			[]entry.Entry{
				{
					Body: "stanza \"log parser\",1,6ft,5",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":   "stanza \"log parser\"",
					"age":    "1",
					"height": "6ft",
					"number": "5",
				},
			},
			false,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewCSVParserConfig("test")
			cfg.OutputIDs = []string{"fake"}
			tc.configure(cfg)

			op, err := cfg.Build(testutil.Logger(t))
			if tc.expectBuildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			fake := testutil.NewFakeOutput(t)
			op.SetOutputs([]operator.Operator{fake})

			for i, inputEntry := range tc.inputEntry {
				err = op.Process(context.Background(), &inputEntry)
				if tc.expectProcessErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				fake.ExpectBody(t, tc.outputBody[i])
			}
		})
	}
}

func TestParserCSVMultiline(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			"first_field",
			"aa\naa,bbbb,cccc,dddd,eeee",
			map[string]interface{}{
				"A": "aa\naa",
				"B": "bbbb",
				"C": "cccc",
				"D": "dddd",
				"E": "eeee",
			},
		},
		{
			"middle_field",
			"aaaa,bbbb,cc\ncc,dddd,eeee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb",
				"C": "cc\ncc",
				"D": "dddd",
				"E": "eeee",
			},
		},
		{
			"last_field",
			"aaaa,bbbb,cccc,dddd,e\neee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb",
				"C": "cccc",
				"D": "dddd",
				"E": "e\neee",
			},
		},
		{
			"multiple_fields",
			"aaaa,bb\nbb,ccc\nc,dddd,e\neee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bb\nbb",
				"C": "ccc\nc",
				"D": "dddd",
				"E": "e\neee",
			},
		},
		{
			"multiple_first_field",
			"a\na\na\na,bbbb,cccc,dddd,eeee",
			map[string]interface{}{
				"A": "a\na\na\na",
				"B": "bbbb",
				"C": "cccc",
				"D": "dddd",
				"E": "eeee",
			},
		},
		{
			"multiple_middle_field",
			"aaaa,bbbb,c\nc\nc\nc,dddd,eeee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb",
				"C": "c\nc\nc\nc",
				"D": "dddd",
				"E": "eeee",
			},
		},
		{
			"multiple_last_field",
			"aaaa,bbbb,cccc,dddd,e\ne\ne\ne",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb",
				"C": "cccc",
				"D": "dddd",
				"E": "e\ne\ne\ne",
			},
		},
		{
			"leading_newline",
			"\naaaa,bbbb,cccc,dddd,eeee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb",
				"C": "cccc",
				"D": "dddd",
				"E": "eeee",
			},
		},
		{
			"trailing_newline",
			"aaaa,bbbb,cccc,dddd,eeee\n",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb",
				"C": "cccc",
				"D": "dddd",
				"E": "eeee",
			},
		},
		{
			"leading_newline_field",
			"aaaa,\nbbbb,\ncccc,\ndddd,eeee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "\nbbbb",
				"C": "\ncccc",
				"D": "\ndddd",
				"E": "eeee",
			},
		},
		{
			"trailing_newline_field",
			"aaaa,bbbb\n,cccc\n,dddd\n,eeee",
			map[string]interface{}{
				"A": "aaaa",
				"B": "bbbb\n",
				"C": "cccc\n",
				"D": "dddd\n",
				"E": "eeee",
			},
		},
		{
			"empty_lines",
			"aa\n\naa,bbbb,c\n\nccc,dddd,eee\n\ne",
			map[string]interface{}{
				"A": "aa\naa",
				"B": "bbbb",
				"C": "c\nccc",
				"D": "dddd",
				"E": "eee\ne",
			},
		},
		{
			"everything",
			"\n\na\na\n\naa,\n\nbb\nbb\n\n,cc\ncc\n\n,\ndddd\n,eeee\n\n",
			map[string]interface{}{
				"A": "a\na\naa",
				"B": "\nbb\nbb\n",
				"C": "cc\ncc\n",
				"D": "\ndddd\n",
				"E": "eeee",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewCSVParserConfig("test")
			cfg.OutputIDs = []string{"fake"}
			cfg.Header = "A,B,C,D,E"

			op, err := cfg.Build(testutil.Logger(t))
			require.NoError(t, err)

			fake := testutil.NewFakeOutput(t)
			op.SetOutputs([]operator.Operator{fake})

			entry := entry.New()
			entry.Body = tc.input
			err = op.Process(context.Background(), entry)
			require.NoError(t, err)
			fake.ExpectBody(t, tc.expected)
			fake.ExpectNoEntry(t, 100*time.Millisecond)
		})
	}
}

func TestParserCSVInvalidJSONInput(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cfg := NewCSVParserConfig("test")
		cfg.OutputIDs = []string{"fake"}
		cfg.Header = testHeader

		op, err := cfg.Build(testutil.Logger(t))
		require.NoError(t, err)

		fake := testutil.NewFakeOutput(t)
		op.SetOutputs([]operator.Operator{fake})

		entry := entry.New()
		entry.Body = "{\"name\": \"stanza\"}"
		err = op.Process(context.Background(), entry)
		require.Error(t, err, "parse error on line 1, column 1: bare \" in non-quoted-field")
		fake.ExpectBody(t, "{\"name\": \"stanza\"}")
	})
}

func TestBuildParserCSV(t *testing.T) {
	newBasicCSVParser := func() *CSVParserConfig {
		cfg := NewCSVParserConfig("test")
		cfg.OutputIDs = []string{"test"}
		cfg.Header = "name,position,number"
		cfg.FieldDelimiter = ","
		return cfg
	}

	t.Run("BasicConfig", func(t *testing.T) {
		c := newBasicCSVParser()
		_, err := c.Build(testutil.Logger(t))
		require.NoError(t, err)
	})

	t.Run("MissingHeaderField", func(t *testing.T) {
		c := newBasicCSVParser()
		c.Header = ""
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
	})

	t.Run("InvalidHeaderFieldMissingDelimiter", func(t *testing.T) {
		c := newBasicCSVParser()
		c.Header = "name"
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing field delimiter in header")
	})

	t.Run("InvalidHeaderFieldWrongDelimiter", func(t *testing.T) {
		c := newBasicCSVParser()
		c.Header = "name;position;number"
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
	})

	t.Run("InvalidDelimiter", func(t *testing.T) {
		c := newBasicCSVParser()
		c.Header = "name,position,number"
		c.FieldDelimiter = ":"
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing field delimiter in header")
	})
}
