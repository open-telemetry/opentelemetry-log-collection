module github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/output/newrelic

go 1.14

require (
	github.com/open-telemetry/opentelemetry-log-collection v0.13.12
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
)

replace github.com/open-telemetry/opentelemetry-log-collection => ../../../../
