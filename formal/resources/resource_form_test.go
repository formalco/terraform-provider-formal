package resource

import (
	"encoding/json"
	"maps"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestExpandTerraformFormFieldConfigOptionsSourceInputJSON(t *testing.T) {
	config, err := expandTerraformFormFieldConfig(formFieldConfigWithOptionsSource(map[string]any{
		"input_json": `{
			"limit": 500,
			"order": "desc",
			"searchFields": ["name", "technology", "hostname"],
			"exact": true
		}`,
	}), 0)
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}

	inputMap := config.OptionsSource.Input.AsMap()
	if got := inputMap["limit"]; got != float64(500) {
		t.Fatalf("expected numeric limit 500, got %#v", got)
	}

	if got := inputMap["exact"]; got != true {
		t.Fatalf("expected boolean exact true, got %#v", got)
	}

	searchFields, ok := inputMap["searchFields"].([]any)
	if !ok {
		t.Fatalf("expected searchFields to be an array, got %#v", inputMap["searchFields"])
	}

	if got := searchFields[1]; got != "technology" {
		t.Fatalf("expected searchFields[1] to be technology, got %#v", got)
	}
}

func TestExpandTerraformFormFieldConfigRejectsInputAndInputJSON(t *testing.T) {
	_, err := expandTerraformFormFieldConfig(formFieldConfigWithOptionsSource(map[string]any{
		"input": map[string]any{
			"limit": "500",
		},
		"input_json": `{"limit":500}`,
	}), 0)
	if err == nil {
		t.Fatalf("expected error when input and input_json are both configured")
	}
}

func TestNormalizeProtoFormFieldConfigUsesInputJSONForNativeValues(t *testing.T) {
	input, err := structpb.NewStruct(map[string]any{
		"limit":        500,
		"order":        "desc",
		"searchFields": []any{"name", "technology", "hostname"},
	})
	if err != nil {
		t.Fatalf("unexpected struct creation error: %v", err)
	}

	config := normalizeProtoFormFieldConfig(&corev1.FormFieldConfig{
		OptionsSource: &corev1.FormFieldOptionsSource{
			App:           "Resource",
			MachineUserId: "machine-user-id",
			Transform:     "body.resources",
			Command: &corev1.FormFieldOptionsSourceCommand{
				Name: "Resources",
			},
			Input: input,
		},
	})

	optionsSource := config["options_source"].([]any)[0].(map[string]any)
	if _, exists := optionsSource["input"]; exists {
		t.Fatalf("expected non-string payload to be flattened as input_json, got input: %#v", optionsSource["input"])
	}

	inputJSON, ok := optionsSource["input_json"].(string)
	if !ok {
		t.Fatalf("expected input_json string, got %#v", optionsSource["input_json"])
	}

	var inputMap map[string]any
	if err := json.Unmarshal([]byte(inputJSON), &inputMap); err != nil {
		t.Fatalf("expected valid input_json, got error: %v", err)
	}

	if got := inputMap["limit"]; got != float64(500) {
		t.Fatalf("expected numeric limit 500, got %#v", got)
	}
}

func TestNormalizeProtoFormFieldConfigPreservesInputMap(t *testing.T) {
	input, err := structpb.NewStruct(map[string]any{
		"limit": "500",
		"order": "desc",
	})
	if err != nil {
		t.Fatalf("unexpected struct creation error: %v", err)
	}

	tfConfig := formFieldConfigWithOptionsSource(map[string]any{
		"input": map[string]any{
			"limit": "500",
			"order": "desc",
		},
	})

	config := normalizeProtoFormFieldConfig(&corev1.FormFieldConfig{
		OptionsSource: &corev1.FormFieldOptionsSource{
			App:           "Resource",
			MachineUserId: "machine-user-id",
			Transform:     "body.resources",
			Command: &corev1.FormFieldOptionsSourceCommand{
				Name: "Resources",
			},
			Input: input,
		},
	}, tfConfig)

	optionsSource := config["options_source"].([]any)[0].(map[string]any)
	if _, exists := optionsSource["input_json"]; exists {
		t.Fatalf("expected string payload to keep using input, got input_json: %#v", optionsSource["input_json"])
	}

	inputMap, ok := optionsSource["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected input map, got %#v", optionsSource["input"])
	}

	if got := inputMap["limit"]; got != "500" {
		t.Fatalf("expected string limit 500, got %#v", got)
	}
}

func TestNormalizeProtoFormFieldConfigPreservesAllStringInputJSON(t *testing.T) {
	tfConfig := formFieldConfigWithOptionsSource(map[string]any{
		"input_json": `{"search":"hello","order":"desc"}`,
	})

	config, err := expandTerraformFormFieldConfig(tfConfig, 0)
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}

	normalizedConfig := normalizeProtoFormFieldConfig(config, tfConfig)
	optionsSource := normalizedConfig["options_source"].([]any)[0].(map[string]any)
	if _, exists := optionsSource["input"]; exists {
		t.Fatalf("expected all-string input_json payload to stay input_json, got input: %#v", optionsSource["input"])
	}

	inputJSON, ok := optionsSource["input_json"].(string)
	if !ok {
		t.Fatalf("expected input_json string, got %#v", optionsSource["input_json"])
	}

	var inputMap map[string]any
	if err := json.Unmarshal([]byte(inputJSON), &inputMap); err != nil {
		t.Fatalf("expected valid input_json, got error: %v", err)
	}

	if got := inputMap["search"]; got != "hello" {
		t.Fatalf("expected search hello, got %#v", got)
	}
}

func TestFormFieldRulesRoundTrip(t *testing.T) {
	tfFields := []any{
		map[string]any{
			"id":   "field_env",
			"name": "Environment",
			"type": "radio",
			"config": []any{map[string]any{
				"option": []any{
					map[string]any{"label": "Dev", "value": "dev"},
					map[string]any{"label": "Prod", "value": "prod"},
				},
			}},
		},
		map[string]any{
			"id":         "field_ticket",
			"name":       "Ticket",
			"type":       "string",
			"visible_if": `field_env == "prod"`,
			"validation": []any{map[string]any{
				"rule":    `field_ticket.matches("^INC-[0-9]+$")`,
				"message": "Use an INC- ticket",
			}},
		},
		map[string]any{
			"id":       "field_duration",
			"name":     "Duration",
			"type":     "select",
			"optional": true,
			"config": []any{map[string]any{
				"option": []any{
					map[string]any{"label": "1 hour", "value": "1h"},
					map[string]any{"label": "24 hours", "value": "24h", "visible_if": `field_env == "dev"`},
				},
			}},
		},
	}

	fields, err := expandTerraformFormFields(tfFields)
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}

	if fields[0].Optional || fields[0].VisibleIf != "" || fields[0].Validation != nil {
		t.Fatalf("expected field_env to have no rules, got %#v", fields[0])
	}
	if got := fields[1].VisibleIf; got != `field_env == "prod"` {
		t.Fatalf("expected field_ticket visible_if, got %q", got)
	}
	if got := fields[1].GetValidation().GetMessage(); got != "Use an INC- ticket" {
		t.Fatalf("expected field_ticket validation message, got %q", got)
	}
	if !fields[2].Optional {
		t.Fatalf("expected field_duration to be optional")
	}
	if got := fields[2].Config.Options[1].VisibleIf; got != `field_env == "dev"` {
		t.Fatalf("expected 24h option visible_if, got %q", got)
	}

	normalized := normalizeProtoFormFields(fields, tfFields)

	ticket := normalized[1].(map[string]any)
	if got := ticket["validation"].([]any)[0].(map[string]any)["rule"]; got != `field_ticket.matches("^INC-[0-9]+$")` {
		t.Fatalf("expected field_ticket rule to round-trip, got %#v", got)
	}
	if _, exists := normalized[0].(map[string]any)["validation"]; exists {
		t.Fatalf("expected no validation block for field_env")
	}

	duration := normalized[2].(map[string]any)
	if duration["optional"] != true {
		t.Fatalf("expected field_duration optional to round-trip")
	}
	option := duration["config"].([]any)[0].(map[string]any)["option"].([]any)[1].(map[string]any)
	if got := option["visible_if"]; got != `field_env == "dev"` {
		t.Fatalf("expected 24h option visible_if to round-trip, got %#v", got)
	}
}

func formFieldConfigWithOptionsSource(optionsSource map[string]any) map[string]any {
	baseOptionsSource := map[string]any{
		"app":             "Resource",
		"machine_user_id": "machine-user-id",
		"transform":       "body.resources",
		"command": []any{
			map[string]any{
				"name": "Resources",
			},
		},
	}

	maps.Copy(baseOptionsSource, optionsSource)

	return map[string]any{
		"options_source": []any{baseOptionsSource},
	}
}
