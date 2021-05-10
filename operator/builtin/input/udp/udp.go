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

package udp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

const (
	// DefaultMaxLogSize is the max buffer sized used
	// if MaxLogSize is not set
	DefaultMaxLogSize = 1024 * 1024
)

func init() {
	operator.Register("udp_input", func() operator.Builder { return NewUDPInputConfig("") })
}

// NewUDPInputConfig creates a new UDP input config with default values
func NewUDPInputConfig(operatorID string) *UDPInputConfig {
	return &UDPInputConfig{
		InputConfig: helper.NewInputConfig(operatorID, "udp_input"),
		Encoding:    helper.NewEncodingConfig(),
		Multiline: helper.MultilineConfig{
			LineStartPattern: "",
			LineEndPattern:   ".^", // Use never matching regex to not split data by default
		},
	}
}

// UDPInputConfig is the configuration of a udp input operator.
type UDPInputConfig struct {
	helper.InputConfig `yaml:",inline"`

	ListenAddress string                 `mapstructure:"listen_address,omitempty"        json:"listen_address,omitempty"       yaml:"listen_address,omitempty"`
	AddAttributes bool                   `mapstructure:"add_attributes,omitempty"        json:"add_attributes,omitempty"       yaml:"add_attributes,omitempty"`
	MaxLogSize    helper.ByteSize        `mapstructure:"max_log_size,omitempty"          json:"max_log_size,omitempty"         yaml:"max_log_size,omitempty"`
	Encoding      helper.EncodingConfig  `mapstructure:",squash,omitempty"               json:",inline,omitempty"              yaml:",inline,omitempty"`
	Multiline     helper.MultilineConfig `mapstructure:"multiline,omitempty"             json:"multiline,omitempty"            yaml:"multiline,omitempty"`
}

// Build will build a udp input operator.
func (c UDPInputConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	// If MaxLogSize not set, set sane default in order to remain
	// backwards compatible with existing plugins and configurations
	if c.MaxLogSize == 0 {
		c.MaxLogSize = DefaultMaxLogSize
	}

	if c.ListenAddress == "" {
		return nil, fmt.Errorf("missing required parameter 'listen_address'")
	}

	address, err := net.ResolveUDPAddr("udp", c.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve listen_address: %s", err)
	}

	encoding, err := c.Encoding.Build(context)
	if err != nil {
		return nil, err
	}

	splitFunc, err := c.Multiline.Build(context, encoding.Encoding, true)
	if err != nil {
		return nil, err
	}

	udpInput := &UDPInput{
		InputOperator: inputOperator,
		address:       address,
		buffer:        make([]byte, DefaultMaxLogSize),
		addAttributes: c.AddAttributes,
		MaxLogSize:    int(c.MaxLogSize),
		encoding:      encoding,
		splitFunc:     splitFunc,
	}
	return []operator.Operator{udpInput}, nil
}

// UDPInput is an operator that listens to a socket for log entries.
type UDPInput struct {
	buffer []byte
	helper.InputOperator
	address       *net.UDPAddr
	addAttributes bool

	connection net.PacketConn
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	MaxLogSize int
	encoding   helper.Encoding
	splitFunc  bufio.SplitFunc
}

// Start will start listening for messages on a socket.
func (u *UDPInput) Start(persister operator.Persister) error {
	ctx, cancel := context.WithCancel(context.Background())
	u.cancel = cancel

	conn, err := net.ListenUDP("udp", u.address)
	if err != nil {
		return fmt.Errorf("failed to open connection: %s", err)
	}
	u.connection = conn

	u.goHandleMessages(ctx)
	return nil
}

// goHandleMessages will handle messages from a udp connection.
func (u *UDPInput) goHandleMessages(ctx context.Context) {
	u.wg.Add(1)

	go func() {
		defer u.wg.Done()

		for {
			message, remoteAddr, err := u.readMessage()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					u.Errorw("Failed reading messages", zap.Error(err))
				}
				break
			}

			// Initial buffer size is 64k
			buf := make([]byte, 0, 64*1024)
			scanner := bufio.NewScanner(bytes.NewReader(message))
			scanner.Buffer(buf, u.MaxLogSize*1024)

			scanner.Split(u.splitFunc)

			for scanner.Scan() {
				decoded, err := u.encoding.Decode(scanner.Bytes())
				if err != nil {
					u.Errorw("Failed to decode data", zap.Error(err))
					continue
				}

				entry, err := u.NewEntry(decoded)
				if err != nil {
					u.Errorw("Failed to create entry", zap.Error(err))
					continue
				}

				if u.addAttributes {
					entry.AddAttribute("net.transport", "IP.UDP")
					if addr, ok := u.connection.LocalAddr().(*net.UDPAddr); ok {
						entry.AddAttribute("net.host.ip", addr.IP.String())
						entry.AddAttribute("net.host.port", strconv.FormatInt(int64(addr.Port), 10))
					}

					if addr, ok := remoteAddr.(*net.UDPAddr); ok {
						entry.AddAttribute("net.peer.ip", addr.IP.String())
						entry.AddAttribute("net.peer.port", strconv.FormatInt(int64(addr.Port), 10))
					}
				}

				u.Write(ctx, entry)
			}
		}
	}()
}

// readMessage will read log messages from the connection.
func (u *UDPInput) readMessage() ([]byte, net.Addr, error) {
	n, addr, err := u.connection.ReadFrom(u.buffer)
	if err != nil {
		return nil, nil, err
	}

	// Remove trailing characters and NULs
	for ; (n > 0) && (u.buffer[n-1] < 32); n-- {
	}

	return u.buffer[:n], addr, nil
}

// Stop will stop listening for udp messages.
func (u *UDPInput) Stop() error {
	u.cancel()
	u.connection.Close()
	u.wg.Wait()
	return nil
}
