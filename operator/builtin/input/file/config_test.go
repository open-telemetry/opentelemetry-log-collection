package file

import (
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/stretchr/testify/require"
	"testing"
)

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
	except := NewTestInputConfig()
	input := map[string]interface{}{
		// InputConfig
		"id":       "config_test",
		"type":     "file_input",
		"write_to": "$",
		"labels": map[string]interface{}{
		},
		"resource": map[string]interface{}{
		},

		"include":       except.Include,
		"exclude":       except.Exclude,
		"poll_interval": "0.2",
		"multiline": map[string]interface{}{
			"line_start_pattern": except.Multiline.LineStartPattern,
			"line_end_pattern":   except.Multiline.LineEndPattern,
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
	require.Equal(t, except, &actual)
}

func TestMapStructureDecodeConfig(t *testing.T) {
	except := NewTestInputConfig()
	input := map[string]interface{}{
		// InputConfig
		"id":   "config_test",
		"type": "file_input",
		"write_to": entry.NewRecordField([]string{}...),
		"labels": map[string]interface{}{
		},
		"resource": map[string]interface{}{
		},
		"include": except.Include,
		"exclude": except.Exclude,
		"poll_interval": map[string]interface{}{
			"Duration": 200 * 1000 * 1000,
		},
		"multiline": map[string]interface{}{
			"line_start_pattern": except.Multiline.LineStartPattern,
			"line_end_pattern":   except.Multiline.LineEndPattern,
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
	require.Equal(t, except, &actual)
}
