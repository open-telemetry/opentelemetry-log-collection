package helper

import (
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/collector/config/configtls"
)

type TLSServerConfig struct {
	*configtls.TLSServerSetting
}

func (t *TLSServerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tlsConfig map[string]interface{}
	err := unmarshal(&tlsConfig)
	if err != nil {
		return err
	}
	return mapstructure.Decode(tlsConfig, &t.TLSServerSetting)
}
