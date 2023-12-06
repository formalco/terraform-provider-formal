package validation

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func IntegrationDatahubSyncEntities() schema.SchemaValidateFunc {
	return func(val interface{}, key string) ([]string, []error) {
		var errs []error
		var warns []string
		validEntities := []string{"tags", "data_labels"}
		entities, ok := val.([]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("expected a list of strings for %s, got: %T", key, val))
			return warns, errs
		}

		for i, entity := range entities {
			if _, ok := entity.(string); !ok {
				errs = append(errs, fmt.Errorf("expected string value for element %d in %s", i, key))
				continue
			}

			if !slices.Contains(validEntities, entity.(string)) {
				errs = append(errs, fmt.Errorf("%q in %s at position %d is not a valid entity", entity, key, i))
			}
		}

		return warns, errs
	}
}
