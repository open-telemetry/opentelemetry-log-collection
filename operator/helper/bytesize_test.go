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

package helper

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type testCase struct {
	input       string
	expected    ByteSize
	expectError bool
}

var sharedTestCases = []testCase{
	{`1`, 1, false},
	{`3.3`, 3, false},
	{`0`, 0, false},
	{`10101010`, 10101010, false},
	{`0.01`, 0, false},
	{`"1"`, 1, false},
	{`"1kb"`, 1000, false},
	{`"1KB"`, 1000, false},
	{`"1kib"`, 1024, false},
	{`"1KiB"`, 1024, false},
	{`"1mb"`, 1000 * 1000, false},
	{`"1mib"`, 1024 * 1024, false},
	{`"1gb"`, 1000 * 1000 * 1000, false},
	{`"1gib"`, 1024 * 1024 * 1024, false},
	{`"1tb"`, 1000 * 1000 * 1000 * 1000, false},
	{`"1tib"`, 1024 * 1024 * 1024 * 1024, false},
	{`"1pB"`, 1000 * 1000 * 1000 * 1000 * 1000, false},
	{`"1pib"`, 1024 * 1024 * 1024 * 1024 * 1024, false},
	{`"3ii3"`, 0, true},
	{`3ii3`, 0, true},
	{`--ii3`, 0, true},
	{`{"test":"val"}`, 0, true},
	// {`1e3`, 1000, false},   not supported in mapstructure
}

func TestByteSizeUnmarshalJSON(t *testing.T) {
	for _, tc := range sharedTestCases {
		t.Run("json"+tc.input, func(t *testing.T) {
			var bs ByteSize
			err := json.Unmarshal([]byte(tc.input), &bs)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, bs)
		})
	}
}

func TestByteSizeUnmarshalYAML(t *testing.T) {
	additionalCases := []testCase{
		{`1kb`, 1000, false},
		{`1KB`, 1000, false},
		{`1kib`, 1024, false},
		{`1KiB`, 1024, false},
		{`1mb`, 1000 * 1000, false},
		{`1mib`, 1024 * 1024, false},
		{`1gb`, 1000 * 1000 * 1000, false},
		{`1gib`, 1024 * 1024 * 1024, false},
		{`1tb`, 1000 * 1000 * 1000 * 1000, false},
		{`1tib`, 1024 * 1024 * 1024 * 1024, false},
		{`1pB`, 1000 * 1000 * 1000 * 1000 * 1000, false},
		{`1pib`, 1024 * 1024 * 1024 * 1024 * 1024, false},
		{`test: val`, 0, true},
	}

	cases := append(sharedTestCases, additionalCases...)

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var bs ByteSize
			yamlErr := yaml.Unmarshal([]byte(tc.input), &bs)
			if tc.expectError {
				require.Error(t, yamlErr)
				return
			}
			require.NoError(t, yamlErr)
			require.Equal(t, tc.expected, bs)
		})
		t.Run("mapstructure"+tc.input, func(t *testing.T) {
			// And also same result if using mapstructure
			var newBs ByteSize
			var raw string
			var savedErr error
			err := yaml.Unmarshal([]byte(tc.input), &raw)

			//saving the error so we can check multiple errors and see if there were any errors at any point
			if err != nil {
				savedErr = err
			}

			dc := &mapstructure.DecoderConfig{Result: &newBs, DecodeHook: JSONUnmarshalerHook()}
			ms, err := mapstructure.NewDecoder(dc)

			if err != nil {
				savedErr = err
			}

			if tc.expectError {
				require.Error(t, savedErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, ms.Decode(raw))

			require.Equal(t, tc.expected, newBs)
		})
	}
}
