package flatten

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/errors"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("flatten", func() operator.Builder { return NewFlattenOperatorConfig("") })
}

// NewFlattenOperatorConfig creates a new restructure operator config with default values
func NewFlattenOperatorConfig(operatorID string) *FlattenOperatorConfig {
	return &FlattenOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "move"),
	}
}

// FlattenOperatorConfig is the configuration of a restructure operator
type FlattenOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	Field                    entry.RecordField `json:"field" yaml:"from,field"`
}

func (c FlattenOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	addOperator := &FlattenOperator{
		TransformerOperator: transformerOperator,
		Field:               c.Field,
	}

	return []operator.Operator{addOperator}, nil
}

type FlattenOperator struct {
	helper.TransformerOperator `mapstructure:",squash" yaml:",inline"`
	Field                      entry.RecordField
}

// Process will process an entry with a restructure transformation.
func (p *FlattenOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *FlattenOperator) Transform(entry *entry.Entry) error {
	parent := p.Field.Parent()
	val, ok := entry.Delete(p.Field)
	if !ok {
		// The field doesn't exist, so ignore it
		return fmt.Errorf("apply flatten: field %s does not exist on record", p.Field)
	}

	valMap, ok := val.(map[string]interface{})
	if !ok {
		// The field we were asked to flatten was not a map, so put it back
		err := entry.Set(p.Field, val)
		if err != nil {
			return errors.Wrap(err, "reset non-map field")
		}
		return fmt.Errorf("apply flatten: field %s is not a map", p.Field)
	}

	for k, v := range valMap {
		err := entry.Set(parent.Child(k), v)
		if err != nil {
			return err
		}
	}
	return nil
}
