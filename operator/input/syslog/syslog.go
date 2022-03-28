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

package syslog

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/input/tcp"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/input/udp"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/parser/syslog"
)

func init() {
	operator.Register("syslog_input", func() operator.Builder { return NewSyslogInputConfig("") })
}
func NewSyslogInputConfig(operatorID string) *SyslogInputConfig {
	return &SyslogInputConfig{
		InputConfig: helper.NewInputConfig(operatorID, "syslog_input"),
	}
}

type SyslogInputConfig struct {
	helper.InputConfig      `yaml:",inline"`
	syslog.SyslogBaseConfig `yaml:",inline"`
	Tcp                     *tcp.TCPBaseConfig `json:"tcp" yaml:"tcp"`
	Udp                     *udp.UDPBaseConfig `json:"udp" yaml:"udp"`
}

func (c SyslogInputConfig) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	inputBase, err := c.InputConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	syslogParserCfg := syslog.NewSyslogParserConfig(inputBase.ID() + "_internal_tcp")
	syslogParserCfg.SyslogBaseConfig = c.SyslogBaseConfig
	syslogParserCfg.SetID(inputBase.ID() + "_internal_parser")
	syslogParserCfg.OutputIDs = c.OutputIDs
	syslogParser, err := syslogParserCfg.Build(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve syslog config: %s", err)
	}

	if c.Tcp != nil {
		tcpInputCfg := tcp.NewTCPInputConfig(inputBase.ID() + "_internal_tcp")
		tcpInputCfg.TCPBaseConfig = *c.Tcp

		tcpInput, err := tcpInputCfg.Build(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tcp config: %s", err)
		}

		tcpInput.SetOutputIDs([]string{syslogParser.ID()})
		if err := tcpInput.SetOutputs([]operator.Operator{syslogParser}); err != nil {
			return nil, fmt.Errorf("failed to set outputs")
		}

		return &SyslogInput{
			InputOperator: inputBase,
			tcp:           tcpInput.(*tcp.TCPInput),
			parser:        syslogParser.(*syslog.SyslogParser),
		}, nil
	}

	if c.Udp != nil {
		udpInputCfg := udp.NewUDPInputConfig(inputBase.ID() + "_internal_udp")
		udpInputCfg.UDPBaseConfig = *c.Udp

		udpInput, err := udpInputCfg.Build(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve upd config: %s", err)
		}

		udpInput.SetOutputIDs([]string{syslogParser.ID()})
		if err := udpInput.SetOutputs([]operator.Operator{syslogParser}); err != nil {
			return nil, fmt.Errorf("failed to set outputs")
		}

		return &SyslogInput{
			InputOperator: inputBase,
			udp:           udpInput.(*udp.UDPInput),
			parser:        syslogParser.(*syslog.SyslogParser),
		}, nil
	}

	return nil, fmt.Errorf("need tcp config or udp config")
}

// SyslogInput is an operator that listens for log entries over tcp.
type SyslogInput struct {
	helper.InputOperator
	tcp    *tcp.TCPInput
	udp    *udp.UDPInput
	parser *syslog.SyslogParser
}

// Start will start listening for log entries over tcp or udp.
func (t *SyslogInput) Start(p operator.Persister) error {
	if t.tcp != nil {
		return t.tcp.Start(p)
	}
	return t.udp.Start(p)
}

// Stop will stop listening for messages.
func (t *SyslogInput) Stop() error {
	if t.tcp != nil {
		return t.tcp.Stop()
	}
	return t.udp.Stop()
}

// SetOutputs will set the outputs of the internal syslog parser.
func (t *SyslogInput) SetOutputs(operators []operator.Operator) error {
	t.parser.SetOutputIDs(t.GetOutputIDs())
	return t.parser.SetOutputs(operators)
}
