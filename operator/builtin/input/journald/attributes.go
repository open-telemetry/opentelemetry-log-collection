package journald

import (
	"go.opentelemetry.io/collector/model/semconv/v1.8.0"
)

var (
	// For field definitions, see https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
	defaultAttributeMapping = map[string]string{
		// MESSAGE_ID
		"CODE_FILE": semconv.AttributeCodeFilepath,
		"CODE_LINE": semconv.AttributeCodeLineNumber,
		"CODE_FUNC": semconv.AttributeCodeFunction,
		// ERRNO
		// SYSLOG_FACILITY
		// SYSLOG_IDENTIFIER
		// SYSLOG_TIMESTAMP
		// SYSLOG_RAW
		// DOCUMENTATION
		"TID": semconv.AttributeThreadID,
	}

	defaultResourceMapping = map[string]string{
		// INVOCATION_ID
		// USER_INVOCATION_ID
		// SYSLOG_PID
		"_PID": semconv.AttributeProcessPID,
		// _UID
		// _GID
		"_COMM":    semconv.AttributeProcessCommand,
		"_EXE":     semconv.AttributeProcessExecutablePath,
		"_CMDLINE": semconv.AttributeProcessCommandLine,
		// _CAP_EFFECTIVE
		// _AUDIT_SESSION
		// _AUDIT_LOGINUID
		// _SYSTEMD_CGROUP
		// _SYSTEMD_SLICE
		// _SYSTEMD_UNIT
		// _SYSTEMD_USER_UNIT
		// _SYSTEMD_USER_SLICE
		// _SYSTEMD_SESSION
		// _SYSTEMD_OWNER_UID
		// _SELINUX_CONTEXT
		// _SOURCE_REALTIME_TIMESTAMP
		// _BOOT_ID
		// _MACHINE_ID
		"_HOSTNAME": semconv.AttributeHostName,
		// _TRANSPORT
		// _STREAM_ID
		// _LINE_BREAK
		// _NAMESPACE
	}
)

func hasFieldMapping(mapping map[string]string, field string) bool {
	_, ok := mapping[field]
	return ok
}
