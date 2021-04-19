package operatortest

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"gopkg.in/yaml.v2"
)

func ConfigFromFileViaYaml(file string, config interface{}) (interface{}, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not find config file: %s", err)
	}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("failed to read config file as yaml: %s", err)
	}

	return config, nil
}

func ConfigFromFileViaMapstructure(file string, config interface{}) (interface{}, error) {
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
