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

package tcp

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

const (
	// minBufferSize is the initial size used for buffering
	// TCP input
	minBufferSize = 64 * 1024

	// DefaultMaxBufferSize is the max buffer sized used
	// if MaxBufferSize is not set
	DefaultMaxBufferSize = 1024 * 1024
)

func init() {
	operator.Register("tcp_input", func() operator.Builder { return NewTCPInputConfig("") })
}

// NewTCPInputConfig creates a new TCP input config with default values
func NewTCPInputConfig(operatorID string) *TCPInputConfig {
	return &TCPInputConfig{
		InputConfig: helper.NewInputConfig(operatorID, "tcp_input"),
	}
}

// TCPInputConfig is the configuration of a tcp input operator.
type TCPInputConfig struct {
	helper.InputConfig `yaml:",inline"`

	MaxBufferSize helper.ByteSize         `json:"max_buffer_size,omitempty" yaml:"max_buffer_size,omitempty"`
	ListenAddress string                  `json:"listen_address,omitempty" yaml:"listen_address,omitempty"`
	TLS           *helper.TLSServerConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
}

// Build will build a tcp input operator.
func (c TCPInputConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	// If MaxBufferSize not set, set sane default in order to remain
	// backwards compatible with existing plugins and configurations
	if c.MaxBufferSize == 0 {
		c.MaxBufferSize = DefaultMaxBufferSize
	}

	if c.MaxBufferSize < minBufferSize {
		return nil, fmt.Errorf("invalid value for parameter 'max_buffer_size', must be equal to or greater than %d bytes", minBufferSize)
	}

	if c.ListenAddress == "" {
		return nil, fmt.Errorf("missing required parameter 'listen_address'")
	}

	// validate the input address
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return nil, fmt.Errorf("failed to resolve listen_address: %s", err)
	}

	tcpInput := &TCPInput{
		InputOperator: inputOperator,
		address:       c.ListenAddress,
		maxBufferSize: int(c.MaxBufferSize),
	}

	if c.TLS != nil {
		tcpInput.tls, err = c.TLS.LoadTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	return []operator.Operator{tcpInput}, nil
}

// TCPInput is an operator that listens for log entries over tcp.
type TCPInput struct {
	helper.InputOperator
	address       string
	maxBufferSize int

	listener net.Listener
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	tls      *tls.Config
}

// Start will start listening for log entries over tcp.
func (t *TCPInput) Start() error {
	if err := t.configureListener(); err != nil {
		return fmt.Errorf("failed to listen on interface: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.goListen(ctx)
	return nil
}

func (t *TCPInput) configureListener() error {
	if t.tls == nil {
		listener, err := net.Listen("tcp", t.address)
		if err != nil {
			return fmt.Errorf("failed to configure tcp listener: %w", err)
		}
		t.listener = listener
		return nil
	}

	t.tls.Time = time.Now
	t.tls.Rand = rand.Reader

	listener, err := tls.Listen("tcp", t.address, t.tls)
	if err != nil {
		return fmt.Errorf("failed to configure tls listener: %w", err)
	}

	t.listener = listener
	return nil
}

// goListenn will listen for tcp connections.
func (t *TCPInput) goListen(ctx context.Context) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		for {
			conn, err := t.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					t.Debugw("Listener accept error", zap.Error(err))
				}
			}

			t.Debugf("Received connection: %s", conn.RemoteAddr().String())
			subctx, cancel := context.WithCancel(ctx)
			t.goHandleClose(subctx, conn)
			t.goHandleMessages(subctx, conn, cancel)
		}
	}()
}

// goHandleClose will wait for the context to finish before closing a connection.
func (t *TCPInput) goHandleClose(ctx context.Context, conn net.Conn) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()
		<-ctx.Done()
		t.Debugf("Closing connection: %s", conn.RemoteAddr().String())
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %s", err)
		}
	}()
}

// goHandleMessages will handles messages from a tcp connection.
func (t *TCPInput) goHandleMessages(ctx context.Context, conn net.Conn, cancel context.CancelFunc) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()
		defer cancel()

		// Initial buffer size is 64k
		buf := make([]byte, 0, 64*1024)
		scanner := bufio.NewScanner(conn)
		scanner.Buffer(buf, t.maxBufferSize*1024)
		for scanner.Scan() {
			entry, err := t.NewEntry(scanner.Text())
			if err != nil {
				t.Errorw("Failed to create entry", zap.Error(err))
				continue
			}
			t.Write(ctx, entry)
		}
		if err := scanner.Err(); err != nil {
			t.Errorw("Scanner error", zap.Error(err))
		}
	}()
}

// Stop will stop listening for log entries over TCP.
func (t *TCPInput) Stop() error {
	t.cancel()

	if err := t.listener.Close(); err != nil {
		return err
	}

	t.wg.Wait()
	return nil
}
