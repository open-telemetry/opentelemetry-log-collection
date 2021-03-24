package remove

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("remove", func() operator.Builder { return NewRemoveOperatorConfig("") })
}

// NewRemoveOperatorConfig creates a new restructure operator config with default values
func NewRemoveOperatorConfig(operatorID string) *RemoveOperatorConfig {
	return &RemoveOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "remove"),
	}
}

// RemoveOperatorConfig is the configuration of a restructure operator
type RemoveOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`

	Fields []entry.Field `mapstructure:"fields"  json:"fields" yaml:"fields"`
}

func (c RemoveOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	removeOperator := &RemoveOperator{
		TransformerOperator: transformerOperator,
		Fields:              c.Fields,
	}

	return []operator.Operator{removeOperator}, nil
}

type RemoveOperator struct {
	helper.TransformerOperator
	Fields []entry.Field
}

// Process will process an entry with a restructure transformation.
func (p *RemoveOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *RemoveOperator) Transform(entry *entry.Entry) error {
	if p.Fields != nil {
		for _, field := range p.Fields {
			entry.Delete(field)
		}
	} else {
		return fmt.Errorf("remove: missing required field 'fields'")
	}
	return nil
}
