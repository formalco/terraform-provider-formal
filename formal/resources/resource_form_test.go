package resource

import (
	"encoding/json"
	"maps"
	"testing"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"google.golang.org/protobuf/types/known/structpb"
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
