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

package agent

import (
	"fmt"
	"time"

	"github.com/open-telemetry/opentelemetry-log-collection/database"
	"github.com/open-telemetry/opentelemetry-log-collection/errors"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogAgentBuilder is a construct used to build a log agent
type LogAgentBuilder struct {
	configFiles   []string
	config        *Config
	logger        *zap.SugaredLogger
	pluginDir     string
	namespace     string
	databaseFile  string
	defaultOutput operator.Operator
}

// NewBuilder creates a new LogAgentBuilder
func NewBuilder(logger *zap.SugaredLogger) *LogAgentBuilder {
	return &LogAgentBuilder{
		logger: logger,
	}
}

// WithPluginDir adds the specified plugin directory when building a log agent
func (b *LogAgentBuilder) WithPluginDir(pluginDir string) *LogAgentBuilder {
	b.pluginDir = pluginDir
	return b
}

// WithConfigFiles adds a list of globs to the search path for config files
func (b *LogAgentBuilder) WithConfigFiles(files []string) *LogAgentBuilder {
	b.configFiles = files
	return b
}

// WithConfig builds the agent with a given, pre-built config
func (b *LogAgentBuilder) WithConfig(cfg *Config) *LogAgentBuilder {
	b.config = cfg
	return b
}

// WithDatabaseFile adds the specified database file when building a log agent
func (b *LogAgentBuilder) WithDatabaseFile(databaseFile, namespace string) *LogAgentBuilder {
	b.databaseFile = databaseFile
	b.namespace = namespace
	return b
}

// WithDefaultOutput adds a default output when building a log agent
func (b *LogAgentBuilder) WithDefaultOutput(defaultOutput operator.Operator) *LogAgentBuilder {
	b.defaultOutput = defaultOutput
	return b
}

// Build will build a new log agent using the values defined on the builder
func (b *LogAgentBuilder) Build() (*LogAgent, error) {
	if b.databaseFile != "" && b.namespace == "" {
		return nil, fmt.Errorf("use of database requires namespace")
	}

	db, err := database.OpenDatabase(b.databaseFile)
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}

	if b.pluginDir != "" {
		if errs := plugin.RegisterPlugins(b.pluginDir, operator.DefaultRegistry); len(errs) != 0 {
			b.logger.Errorw("Got errors parsing plugins", "errors", errs)
		}
	}

	if b.config != nil && len(b.configFiles) > 0 {
		return nil, errors.NewError("agent can be built WithConfig or WithConfigFiles, but not both", "")
	}
	if b.config == nil && len(b.configFiles) == 0 {
		return nil, errors.NewError("agent cannot be built without WithConfig or WithConfigFiles", "")
	}
	if len(b.configFiles) > 0 {
		b.config, err = NewConfigFromGlobs(b.configFiles)
		if err != nil {
			return nil, errors.Wrap(err, "read configs from globs")
		}
	}

	sampledLogger := b.logger.Desugar().WithOptions(
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 1, 10000)
		}),
	).Sugar()

	buildContext := operator.NewBuildContext(db, sampledLogger)
	if b.namespace != "" {
		buildContext = buildContext.WithSubNamespace(b.namespace)
	}

	pipeline, err := b.config.Pipeline.BuildPipeline(buildContext, b.defaultOutput)
	if err != nil {
		return nil, err
	}

	return &LogAgent{
		pipeline:      pipeline,
		database:      db,
		SugaredLogger: b.logger,
	}, nil
}
