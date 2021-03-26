package add

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr/vm"
	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

func init() {
	operator.Register("add", func() operator.Builder { return NewAddOperatorConfig("") })
}

// NewAddOperatorConfig creates a new restructure operator config with default values
func NewAddOperatorConfig(operatorID string) *AddOperatorConfig {
	return &AddOperatorConfig{
		TransformerConfig: helper.NewTransformerConfig(operatorID, "add"),
	}
}

// AddOperatorConfig is the configuration of a restructure operator
type AddOperatorConfig struct {
	helper.TransformerConfig `mapstructure:",squash" yaml:",inline"`
	Field                    entry.Field `json:"field" yaml:"field"`
	Value                    interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	program                  *vm.Program
}

func (c AddOperatorConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(context)
	if err != nil {
		return nil, err
	}

	addOperator := &AddOperator{
		TransformerOperator: transformerOperator,
		Field:               c.Field,
		program:             c.program,
		Value:               c.Value,
	}

	return []operator.Operator{addOperator}, nil
}

type AddOperator struct {
	helper.TransformerOperator `mapstructure:",squash" yaml:",inline"`

	Field   entry.Field `json:"field" yaml:"field"`
	Value   interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	program *vm.Program
}

// Process will process an entry with a restructure transformation.
func (p *AddOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the restructure operations to an entry
func (p *AddOperator) Transform(entry *entry.Entry) error {
	if p.Value != nil {
		err := entry.Set(p.Field, p.Value)
		if err != nil {
			return err
		}
	} else if p.program != nil {
		env := helper.GetExprEnv(entry)
		defer helper.PutExprEnv(env)

		result, err := vm.Run(p.program, env)
		if err != nil {
			return fmt.Errorf("evaluate value_expr: %s", err)
		}
		err = entry.Set(p.Field, result)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("add: missing required field 'Value'")
	}
	return nil
}
