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
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type testCase struct {
	name      string
	expectErr bool
	expect    *InputConfig
}

func TestConfig(t *testing.T) {
	cases := []testCase{
		{
			"default",
			false,
			defaultCfg(),
		},
		{
			"id_custom",
			false,
			NewInputConfig("test_id"),
		},
		{
			"include_one",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log")
				return cfg
			}(),
		},
		{
			"include_multi",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "one.log", "two.log", "three.log")
				return cfg
			}(),
		},
		{
			"include_glob",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				return cfg
			}(),
		},
		{
			"include_inline",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "a.log", "b.log")
				return cfg
			}(),
		},
		{
			"exclude_one",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "one.log")
				return cfg
			}(),
		},
		{
			"exclude_multi",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "one.log", "two.log", "three.log")
				return cfg
			}(),
		},
		{
			"exclude_glob",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "not*.log")
				return cfg
			}(),
		},
		{
			"exclude_inline",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.Include = append(cfg.Include, "*.log")
				cfg.Exclude = append(cfg.Exclude, "a.log", "b.log")
				return cfg
			}(),
		},
		{
			"poll_interval_no_units",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Second)
				return cfg
			}(),
		},
		{
			"poll_interval_1s",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Second)
				return cfg
			}(),
		},
		{
			"poll_interval_1ms",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Millisecond)
				return cfg
			}(),
		},
		{
			"poll_interval_1000ms",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.PollInterval = helper.NewDuration(time.Second)
				return cfg
			}(),
		},
		{
			"fingerprint_size_no_units",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1000)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1kb_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1000)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1KB",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1000)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1kib_lower",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1024)
				return cfg
			}(),
		},
		{
			"fingerprint_size_1KiB",
			false,
			func() *InputConfig {
				cfg := defaultCfg()
				cfg.FingerprintSize = helper.ByteSize(1024)
				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFromYaml, err := configFromFileViaYaml(t, path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			require.NoError(t, err)
			require.Equal(t, tc.expect, cfgFromYaml)

			// TODO cfgFromMapstructure, err := configFromFileViaYaml(t, path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.name)))
			// TODO require.NoError(t, err)
			// TODO require.Equal(t, tc.expect, cfgFromMapstructure)
		})
	}
}

func configFromFileViaYaml(t *testing.T, file string) (*InputConfig, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}

	config := NewInputConfig("file_input")
	if err := yaml.UnmarshalStrict(bytes, config); err != nil {
		return nil, fmt.Errorf("failed to read config file as yaml: %s", err)
	}

	return config, nil
}

// TODO func configFromFileViaMapstructure(t *testing.T, file string) (*InputConfig, error) {}

func defaultCfg() *InputConfig {
	return NewInputConfig("file_input")
}


func NewTestInputConfig() *InputConfig {
	cfg := NewInputConfig("config_test")
	cfg.WriteTo = entry.NewRecordField([]string{}...)
	cfg.Include = []string{"i1", "i2"}
	cfg.Exclude = []string{"e1", "e2"}
	cfg.Multiline = &MultilineConfig{"start", "end"}
	cfg.FingerprintSize = 1024
	cfg.Encoding = "utf16"
	return cfg
}

func TestMapStructureDecodeConfigWithHook(t *testing.T) {
	expect := NewTestInputConfig()
	input := map[string]interface{}{
		// InputConfig
		"id":       "config_test",
		"type":     "file_input",
		"write_to": "$",
		"labels": map[string]interface{}{
		},
		"resource": map[string]interface{}{
		},

		"include":       expect.Include,
		"exclude":       expect.Exclude,
		"poll_interval": "0.2",
		"multiline": map[string]interface{}{
			"line_start_pattern": expect.Multiline.LineStartPattern,
			"line_end_pattern":   expect.Multiline.LineEndPattern,
		},
		"include_file_name":    true,
		"include_file_path":    false,
		"start_at":             "end",
		"fingerprint_size":     "1024",
		"max_log_size":         "1mib",
		"max_concurrent_files": 1024,
		"encoding":             "utf16",
	}

	var actual InputConfig
	dc := &mapstructure.DecoderConfig{Result: &actual, DecodeHook: mapstructure.TextUnmarshallerHookFunc()}
	ms, err := mapstructure.NewDecoder(dc)
	require.NoError(t, err)
	err = ms.Decode(input)
	require.NoError(t, err)
	require.Equal(t, expect, &actual)
}

func TestMapStructureDecodeConfig(t *testing.T) {
	expect := NewTestInputConfig()
	input := map[string]interface{}{
		// InputConfig
		"id":   "config_test",
		"type": "file_input",
		"write_to": entry.NewRecordField([]string{}...),
		"labels": map[string]interface{}{
		},
		"resource": map[string]interface{}{
		},
		"include": expect.Include,
		"exclude": expect.Exclude,
		"poll_interval": map[string]interface{}{
			"Duration": 200 * 1000 * 1000,
		},
		"multiline": map[string]interface{}{
			"line_start_pattern": expect.Multiline.LineStartPattern,
			"line_end_pattern":   expect.Multiline.LineEndPattern,
		},
		"include_file_name":    true,
		"include_file_path":    false,
		"start_at":             "end",
		"fingerprint_size":     1024,
		"max_log_size":         1024 * 1024,
		"max_concurrent_files": 1024,
		"encoding":             "utf16",
	}

	var actual InputConfig
	err := mapstructure.Decode(input, &actual)
	require.NoError(t, err)
	require.Equal(t, expect, &actual)
}