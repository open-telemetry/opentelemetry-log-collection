package syslog

import (
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/input/tcp"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/parser/syslog"
	"github.com/open-telemetry/opentelemetry-log-collection/pipeline"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func TestSyslogInput(t *testing.T) {

	t.Run("Simple", SyslogTcpInputTest([]byte("<86>1 2015-08-05T21:58:59.693Z 192.168.2.132 SecureAuth0 23108 ID52020 [SecureAuth@27389 UserHostAddress=\"192.168.2.132\" Realm=\"SecureAuth0\" UserID=\"Tester2\" PEN=\"27389\"] Found the user for retrieving user's profile\n"),
		map[string]interface{}{
			"appname":  "SecureAuth0",
			"facility": 10,
			"hostname": "192.168.2.132",
			"message":  "Found the user for retrieving user's profile",
			"msg_id":   "ID52020",
			"priority": 86,
			"proc_id":  "23108",
			"structured_data": map[string]map[string]string{
				"SecureAuth@27389": {
					"PEN":             "27389",
					"Realm":           "SecureAuth0",
					"UserHostAddress": "192.168.2.132",
					"UserID":          "Tester2",
				},
			},
			"version": 1,
		}))

}

func SyslogTcpInputTest(input []byte, expected interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		cfg := NewSyslogInputConfig("test_syslog")
		cfg.Tcp = tcp.NewTCPInputConfig("test_syslog_tcp")
		cfg.Tcp.ListenAddress = ":22345"
		cfg.Syslog = syslog.NewSyslogParserConfig("test_syslog_parser")
		cfg.Syslog.Protocol = "rfc5424"
		cfg.OutputIDs = []string{"fake"}

		ops, err := cfg.Build(testutil.NewBuildContext(t))
		require.NoError(t, err)
		fake := testutil.NewFakeOutput(t)

		ops = append(ops, fake)
		require.NoError(t, err)
		p, err := pipeline.NewDirectedPipeline(ops)
		require.NoError(t, err)

		err = p.Start()
		require.NoError(t, err)
		defer p.Stop()

		conn, err := net.Dial("tcp", cfg.Tcp.ListenAddress)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write(input)
		require.NoError(t, err)

		select {
		case e := <-fake.Received:
			require.Equal(t, expected, e.Record)
		case <-time.After(time.Second):
			require.FailNow(t, "Timed out waiting for entry to be processed")
		}
	}
}
