// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyvalue

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func newTestParser(t *testing.T) *KVParser {
	config := NewKVParserConfig("test")
	ops, err := config.Build(testutil.NewBuildContext(t))
	op := ops[0]
	require.NoError(t, err)
	return op.(*KVParser)
}

func TestKVParserConfigBuild(t *testing.T) {
	config := NewKVParserConfig("test")
	ops, err := config.Build(testutil.NewBuildContext(t))
	op := ops[0]
	require.NoError(t, err)
	require.IsType(t, &KVParser{}, op)
}

func TestKVParserConfigBuildFailure(t *testing.T) {
	config := NewKVParserConfig("test")
	config.OnError = "invalid_on_error"
	_, err := config.Build(testutil.NewBuildContext(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid `on_error` field")
}

func TestBuild(t *testing.T) {
	basicConfig := func() *KVParserConfig {
		cfg := NewKVParserConfig("test_operator_id")
		return cfg
	}

	cases := []struct {
		name      string
		input     *KVParserConfig
		expectErr bool
	}{
		{
			"default",
			func() *KVParserConfig {
				cfg := basicConfig()
				return cfg
			}(),
			false,
		},
		{
			"delimiter",
			func() *KVParserConfig {
				cfg := basicConfig()
				cfg.Delimiter = "/"
				return cfg
			}(),
			false,
		},
		{
			"missing-delimiter",
			func() *KVParserConfig {
				cfg := basicConfig()
				cfg.Delimiter = ""
				return cfg
			}(),
			true,
		},
		{
			"pair-delimiter",
			func() *KVParserConfig {
				cfg := basicConfig()
				cfg.PairDelimiter = "|"
				return cfg
			}(),
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.input
			_, err := cfg.Build(testutil.NewBuildContext(t))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestKVParserStringFailure(t *testing.T) {
	parser := newTestParser(t)
	_, err := parser.parse("invalid")
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("expected '%s' to split by '%s' into two items, got", "invalid", parser.delimiter))
}

func TestKVParserInvalidType(t *testing.T) {
	parser := newTestParser(t)
	_, err := parser.parse([]int{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "type []int cannot be parsed as key value pairs")
}

func TestKVImplementations(t *testing.T) {
	require.Implements(t, (*operator.Operator)(nil), new(KVParser))
}

func TestKVParser(t *testing.T) {
	cases := []struct {
		name        string
		configure   func(*KVParserConfig)
		inputBody   interface{}
		outputBody  interface{}
		expectError bool
	}{
		{
			"simple",
			func(kv *KVParserConfig) {},
			"name=stanza age=2",
			map[string]interface{}{
				"name": "stanza",
				"age":  "2",
			},
			false,
		},
		{
			"parse-from",
			func(kv *KVParserConfig) {
				kv.ParseFrom = entry.NewBodyField("test")
			},
			map[string]interface{}{
				"test": "name=otel age=2",
			},
			map[string]interface{}{
				"name": "otel",
				"age":  "2",
			},
			false,
		},
		{
			"parse-to",
			func(kv *KVParserConfig) {
				kv.ParseTo = entry.NewBodyField("test")
			},
			"name=stanza age=10",
			map[string]interface{}{
				"test": map[string]interface{}{
					"name": "stanza",
					"age":  "10",
				},
			},
			false,
		},
		{
			"preserve-to",
			func(kv *KVParserConfig) {
				preserveTo := entry.NewBodyField("test")
				kv.PreserveTo = &preserveTo
			},
			"name=stanza age=10",
			map[string]interface{}{
				"name": "stanza",
				"age":  "10",
				"test": "name=stanza age=10",
			},
			false,
		},
		{
			"from-to-preserve",
			func(kv *KVParserConfig) {
				kv.ParseFrom = entry.NewBodyField("from")
				kv.ParseTo = entry.NewBodyField("to")
				orig := entry.NewBodyField("orig")
				kv.PreserveTo = &orig
			},
			map[string]interface{}{
				"from": "name=stanza age=10",
			},
			map[string]interface{}{
				"to": map[string]interface{}{
					"name": "stanza",
					"age":  "10",
				},
				"orig": "name=stanza age=10",
			},
			false,
		},
		{
			"user-agent",
			func(kv *KVParserConfig) {},
			`requestClientApplication="Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.0"`,
			map[string]interface{}{
				"requestClientApplication": `Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.0`,
			},
			false,
		},
		{
			"double-quotes-removed",
			func(kv *KVParserConfig) {},
			"name=\"stanza\" age=2",
			map[string]interface{}{
				"name": "stanza",
				"age":  "2",
			},
			false,
		},
		{
			"double-quotes-spaces-removed",
			func(kv *KVParserConfig) {},
			`name=" stanza " age=2`,
			map[string]interface{}{
				"name": "stanza",
				"age":  "2",
			},
			false,
		},
		{
			"leading-and-trailing-space",
			func(kv *KVParserConfig) {},
			`" name "=" stanza " age=2`,
			map[string]interface{}{
				"name": "stanza",
				"age":  "2",
			},
			false,
		},
		{
			"delimiter",
			func(kv *KVParserConfig) {
				kv.Delimiter = "|"
				kv.ParseFrom = entry.NewBodyField("testfield")
				kv.ParseTo = entry.NewBodyField("testparsed")
			},
			map[string]interface{}{
				"testfield": `name|" stanza " age|2     key|value`,
			},
			map[string]interface{}{
				"testparsed": map[string]interface{}{
					"name": "stanza",
					"age":  "2",
					"key":  "value",
				},
			},
			false,
		},
		{
			"double-delimiter",
			func(kv *KVParserConfig) {
				kv.Delimiter = "=="
			},
			`name==" stanza " age==2     key==value`,
			map[string]interface{}{
				"name": "stanza",
				"age":  "2",
				"key":  "value",
			},
			false,
		},
		{
			"pair-delimiter",
			func(kv *KVParserConfig) {
				kv.PairDelimiter = "|"
			},
			`name=stanza|age=2     | key=value`,
			map[string]interface{}{
				"name": "stanza",
				"age":  "2",
				"key":  "value",
			},
			false,
		},
		{
			"large",
			func(kv *KVParserConfig) {},
			"name=stanza age=1 job=\"software engineering\" location=\"grand rapids michigan\" src=\"10.3.3.76\" dst=172.217.0.10 protocol=udp sport=57112 dport=443 translated_src_ip=96.63.176.3 translated_port=57112",
			map[string]interface{}{
				"age":               "1",
				"dport":             "443",
				"dst":               "172.217.0.10",
				"job":               "software engineering",
				"location":          "grand rapids michigan",
				"name":              "stanza",
				"protocol":          "udp",
				"sport":             "57112",
				"src":               "10.3.3.76",
				"translated_port":   "57112",
				"translated_src_ip": "96.63.176.3",
			},
			false,
		},
		{
			"dell-sonic-wall",
			func(kv *KVParserConfig) {},
			`id=LVM_Sonicwall sn=22255555 time="2021-09-22 16:30:31" fw=14.165.177.10 pri=6 c=1024 gcat=2 m=97 msg="Web site hit" srcMac=6c:0b:84:3f:fa:63 src=192.168.50.2:52006:X0 srcZone=LAN natSrc=14.165.177.10:58457 dstMac=08:b2:58:46:30:54 dst=15.159.150.83:443:X1 dstZone=WAN natDst=15.159.150.83:443 proto=tcp/https sent=1422 rcvd=5993 rule="6 (LAN->WAN)" app=48 dstname=example.space.dev.com arg=/ code=27 Category="Information Technology/Computers" note="Policy: a0, Info: 888 " n=3412158`,
			map[string]interface{}{
				"id":       "LVM_Sonicwall",
				"sn":       "22255555",
				"time":     "2021-09-22 16:30:31",
				"fw":       "14.165.177.10",
				"pri":      "6",
				"c":        "1024",
				"gcat":     "2",
				"m":        "97",
				"msg":      "Web site hit",
				"srcMac":   "6c:0b:84:3f:fa:63",
				"src":      "192.168.50.2:52006:X0",
				"srcZone":  "LAN",
				"natSrc":   "14.165.177.10:58457",
				"dstMac":   "08:b2:58:46:30:54",
				"dst":      "15.159.150.83:443:X1",
				"dstZone":  "WAN",
				"natDst":   "15.159.150.83:443",
				"proto":    "tcp/https",
				"sent":     "1422",
				"rcvd":     "5993",
				"rule":     "6 (LAN->WAN)",
				"app":      "48",
				"dstname":  "example.space.dev.com",
				"arg":      "/",
				"code":     "27",
				"Category": "Information Technology/Computers",
				"note":     "Policy: a0, Info: 888",
				"n":        "3412158",
			},
			false,
		},
		{
			"missing-delimiter",
			func(kv *KVParserConfig) {},
			`test text`,
			nil,
			true,
		},
		{
			"invalid-pair",
			func(kv *KVParserConfig) {},
			`test=text=abc`,
			map[string]interface{}{},
			true,
		},
		{
			"empty-input",
			func(kv *KVParserConfig) {},
			"",
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewKVParserConfig("test")
			cfg.OutputIDs = []string{"fake"}
			tc.configure(cfg)

			ops, err := cfg.Build(testutil.NewBuildContext(t))
			require.NoError(t, err)
			op := ops[0]

			fake := testutil.NewFakeOutput(t)
			op.SetOutputs([]operator.Operator{fake})

			entry := entry.New()
			entry.Body = tc.inputBody
			err = op.Process(context.Background(), entry)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				fake.ExpectBody(t, tc.outputBody)
			}
		})
	}
}

func TestSplitStringByWhitespace(t *testing.T) {
	cases := []struct {
		name   string
		intput string
		output []string
	}{
		{
			"simple",
			"k=v a=b x=\" y \" job=\"software engineering\"",
			[]string{
				"k=v",
				"a=b",
				"x=\" y \"",
				"job=\"software engineering\"",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.output, splitStringByWhitespace(tc.intput))
		})
	}
}
