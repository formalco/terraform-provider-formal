package resource

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
)

// validateCelSyntaxAttribute is the schema validator for attributes holding raw
// CEL, catching typos during plan rather than at apply.
//
// Terraform skips validation for values it does not know yet, so an expression
// built by interpolation is checked once its inputs are known.
func validateCelSyntaxAttribute(value any, key string) ([]string, []error) {
	expression, ok := value.(string)
	if !ok {
		return nil, []error{fmt.Errorf("%s: expected a CEL expression as a string", key)}
	}
	if strings.TrimSpace(expression) == "" {
		return nil, nil
	}
	env, err := cel.NewEnv()
	if err != nil {
		return nil, []error{fmt.Errorf("%s: initialize CEL parser: %w", key, err)}
	}
	if _, issues := env.Parse(expression); issues != nil && issues.Err() != nil {
		return nil, []error{fmt.Errorf("%s is not a valid CEL expression: %w", key, issues.Err())}
	}
	return nil, nil
}
