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

//go:build windows
// +build windows

package windows

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
)

func TestParseValidTimestamp(t *testing.T) {
	xml := EventXML{
		TimeCreated: TimeCreated{
			SystemTime: "2020-07-30T01:01:01.123456789Z",
		},
	}
	timestamp := xml.parseTimestamp()
	expected, _ := time.Parse(time.RFC3339Nano, "2020-07-30T01:01:01.123456789Z")
	require.Equal(t, expected, timestamp)
}

func TestParseInvalidTimestamp(t *testing.T) {
	xml := EventXML{
		TimeCreated: TimeCreated{
			SystemTime: "invalid",
		},
	}
	timestamp := xml.parseTimestamp()
	require.Equal(t, time.Now().Year(), timestamp.Year())
	require.Equal(t, time.Now().Month(), timestamp.Month())
	require.Equal(t, time.Now().Day(), timestamp.Day())
}

func TestParseSeverity(t *testing.T) {
	xmlCritical := EventXML{Level: "Critical"}
	xmlError := EventXML{Level: "Error"}
	xmlWarning := EventXML{Level: "Warning"}
	xmlInformation := EventXML{Level: "Information"}
	xmlUnknown := EventXML{Level: "Unknown"}
	require.Equal(t, entry.Fatal, xmlCritical.parseSeverity())
	require.Equal(t, entry.Error, xmlError.parseSeverity())
	require.Equal(t, entry.Warn, xmlWarning.parseSeverity())
	require.Equal(t, entry.Info, xmlInformation.parseSeverity())
	require.Equal(t, entry.Default, xmlUnknown.parseSeverity())
}

func TestParseBody(t *testing.T) {
	xml := EventXML{
		EventID: EventID{
			ID:         1,
			Qualifiers: 2,
		},
		Provider: Provider{
			Name:            "provider",
			GUID:            "guid",
			EventSourceName: "event source",
		},
		TimeCreated: TimeCreated{
			SystemTime: "2020-07-30T01:01:01.123456789Z",
		},
		Computer:         "computer",
		Channel:          "application",
		RecordID:         1,
		Level:            "Information",
		Message:          "message",
		Task:             "task",
		Opcode:           "opcode",
		Keywords:         []string{"keyword"},
		EventData:        []string{"this", "is", "some", "sample", "data"},
		RenderedLevel:    "rendered_level",
		RenderedTask:     "rendered_task",
		RenderedOpcode:   "rendered_opcode",
		RenderedKeywords: []string{"RenderedKeywords"},
	}

	expected := map[string]interface{}{
		"event_id": map[string]interface{}{
			"id":         uint32(1),
			"qualifiers": uint16(2),
		},
		"provider": map[string]interface{}{
			"name":         "provider",
			"guid":         "guid",
			"event_source": "event source",
		},
		"system_time":       "2020-07-30T01:01:01.123456789Z",
		"computer":          "computer",
		"channel":           "application",
		"record_id":         uint64(1),
		"level":             "Information",
		"message":           "message",
		"task":              "task",
		"opcode":            "opcode",
		"keywords":          []string{"keyword"},
		"rendered_level":    "rendered_level",
		"rendered_task":     "rendered_task",
		"rendered_opcode":   "rendered_opcode",
		"rendered_keywords": []string{"RenderedKeywords"},
		"event_data":        []string{"this", "is", "some", "sample", "data"},
	}

	require.Equal(t, expected, xml.parseBody())
}

func TestParseBody2(t *testing.T) {
	xml := EventXML{
		EventID: EventID{
			ID:         1,
			Qualifiers: 2,
		},
		Provider: Provider{
			Name:            "provider",
			GUID:            "guid",
			EventSourceName: "event source",
		},
		TimeCreated: TimeCreated{
			SystemTime: "2020-07-30T01:01:01.123456789Z",
		},
		Computer:         "computer",
		Channel:          "Security",
		RecordID:         1,
		Level:            "Information",
		Message:          "message",
		Task:             "task",
		Opcode:           "opcode",
		Keywords:         []string{"keyword"},
		EventData:        []string{"this", "is", "some", "sample", "data"},
		RenderedLevel:    "rendered_level",
		RenderedTask:     "rendered_task",
		RenderedOpcode:   "rendered_opcode",
		RenderedKeywords: []string{"RenderedKeywords"},
	}

	expected := map[string]interface{}{
		"event_id": map[string]interface{}{
			"id":         uint32(1),
			"qualifiers": uint16(2),
		},
		"provider": map[string]interface{}{
			"name":         "provider",
			"guid":         "guid",
			"event_source": "event source",
		},
		"system_time":       "2020-07-30T01:01:01.123456789Z",
		"computer":          "computer",
		"channel":           "Security",
		"record_id":         uint64(1),
		"level":             "Information",
		"message":           "message",
		"task":              "task",
		"opcode":            "opcode",
		"keywords":          []string{"keyword"},
		"rendered_level":    "rendered_level",
		"rendered_task":     "rendered_task",
		"rendered_opcode":   "rendered_opcode",
		"rendered_keywords": []string{"RenderedKeywords"},
		"event_data":        []string{"this", "is", "some", "sample", "data"},
	}

	require.Equal(t, expected, xml.parseBody())
}

func TestInvalidUnmarshal(t *testing.T) {
	_, err := unmarshalEventXML([]byte("Test \n Invalid \t Unmarshal"))
	require.Error(t, err)

}
func TestUnmarshal(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "xmlSample.xml"))
	require.NoError(t, err)

	event, err := unmarshalEventXML(data)
	require.NoError(t, err)

	xml := EventXML{
		EventID: EventID{
			ID:         16384,
			Qualifiers: 16384,
		},
		Provider: Provider{
			Name:            "Microsoft-Windows-Security-SPP",
			GUID:            "{E23B33B0-C8C9-472C-A5F9-F2BDFEA0F156}",
			EventSourceName: "Software Protection Platform Service",
		},
		TimeCreated: TimeCreated{
			SystemTime: "2022-04-22T10:20:52.3778625Z",
		},
		Computer:  "computer",
		Channel:   "Application",
		RecordID:  23401,
		Level:     "4",
		Message:   "",
		Task:      "0",
		Opcode:    "0",
		EventData: []string{"2022-04-28T19:48:52Z", "RulesEngine"},
		Keywords:  []string{"0x80000000000000"},
	}

	require.Equal(t, xml, event)
}
