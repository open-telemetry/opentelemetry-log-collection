package operatortest

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// ConfigTestCase is used for testing golden configs
type ConfigTestCase struct {
	Name   string
	Expect interface{}
}

func configFromFileViaYaml(file string, config interface{}) (interface{}, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("failed to read config file as yaml: %s", err)
	}

	return config, nil
}

func configFromFileViaMapstructure(file string, config interface{}) (interface{}, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}

	raw := map[string]interface{}{}

	if err := yaml.Unmarshal(bytes, raw); err != nil {
		return nil, fmt.Errorf("failed to read data from yaml: %s", err)
	}

	dc := &mapstructure.DecoderConfig{Result: config, DecodeHook: helper.JSONUnmarshalerHook()}
	ms, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return nil, err
	}
	err = ms.Decode(raw)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// RunGoldenConfigTest Unmarshalls yaml files and compares them against the expected.
func RunGoldenConfigTest(config interface{}, t *testing.T, tc ConfigTestCase) {
	cfgFromYaml, yamlErr := configFromFileViaYaml(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.Name)), config)
	cfgFromMapstructure, mapErr := configFromFileViaMapstructure(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", tc.Name)), config)
	require.NoError(t, yamlErr)
	require.Equal(t, tc.Expect, cfgFromYaml)
	require.NoError(t, mapErr)
	require.Equal(t, tc.Expect, cfgFromMapstructure)
}
