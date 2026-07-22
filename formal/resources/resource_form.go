package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/protobuf/types/known/structpb"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

var validFormFieldTypes = []string{
	"string",
	"number",
	"timestamp",
	"email",
	"url",
	"date",
	"time",
	"select",
	"multi_select",
	"checkbox",
	"radio",
}

func ResourceForm() *schema.Resource {
	return &schema.Resource{
		Description:   "Forms define reusable input schemas for workflows.",
		CreateContext: resourceFormCreate,
		ReadContext:   resourceFormRead,
		UpdateContext: resourceFormUpdate,
		DeleteContext: resourceFormDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the form.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the form. Must be unique within the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The form description.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"field": {
				Description: "List of fields that define the form schema.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Unique field identifier.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"name": {
							Description: "Display name of the field.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"type": {
							Description: "Field type.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								validFormFieldTypes,
								false,
							),
						},
						"config": {
							Description: "Optional field configuration for select-like field types.",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"option": {
										Description: "Static options for select-like fields.",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"label": {
													Description: "Option label.",
													Type:        schema.TypeString,
													Required:    true,
												},
												"value": {
													Description: "Option value.",
													Type:        schema.TypeString,
													Required:    true,
												},
											},
										},
									},
									"options_source": {
										Description: "Dynamic source used to fetch options.",
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app": {
													Description: "Service/app name used to fetch options.",
													Type:        schema.TypeString,
													Required:    true,
												},
												"command": {
													Description: "Command configuration for options retrieval.",
													Type:        schema.TypeList,
													Required:    true,
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Description: "Command name.",
																Type:        schema.TypeString,
																Required:    true,
															},
														},
													},
												},
												"machine_user_id": {
													Description: "Machine user used to authenticate options retrieval.",
													Type:        schema.TypeString,
													Required:    true,
												},
												"input": {
													Description: "Optional payload for options retrieval.",
													Type:        schema.TypeMap,
													Optional:    true,
												},
												"input_json": {
													Description:  "Optional payload for options retrieval as a JSON object string. Use this when the payload contains non-string JSON values such as numbers, booleans, arrays, or nested objects. Mutually exclusive with input.",
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateJSONObjectString,
													StateFunc:    canonicalizeJSONString,
												},
												"transform": {
													Description: "CEL expression that transforms the response into options.",
													Type:        schema.TypeString,
													Required:    true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"created_at": {
				Description: "When the form was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceFormCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	fields, err := expandTerraformFormFields(d.Get("field").([]any))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.WorkflowServiceClient.CreateForm(ctx, &corev1.CreateFormRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Fields:      fields,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Form.Id)

	return resourceFormRead(ctx, d, meta)
}

func resourceFormRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	formID := d.Id()

	res, err := c.Grpc.Sdk.WorkflowServiceClient.GetForm(ctx, &corev1.GetFormRequest{Id: formID})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Form with ID "+formID+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	form := res.Form

	if err := d.Set("id", form.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", form.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", form.Description); err != nil {
		return diag.FromErr(err)
	}
	currentFields, _ := d.Get("field").([]any)
	if err := d.Set("field", normalizeProtoFormFields(form.Fields, currentFields)); err != nil {
		return diag.FromErr(err)
	}
	if form.CreatedAt != nil {
		if err := d.Set("created_at", form.CreatedAt.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}
	if form.UpdatedAt != nil {
		if err := d.Set("updated_at", form.UpdatedAt.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceFormUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("field") {
		fields, err := expandTerraformFormFields(d.Get("field").([]any))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = c.Grpc.Sdk.WorkflowServiceClient.UpdateForm(ctx, &corev1.UpdateFormRequest{
			Id:          d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Fields:      fields,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFormRead(ctx, d, meta)
}

func resourceFormDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.WorkflowServiceClient.DeleteForm(ctx, &corev1.DeleteFormRequest{Id: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandTerraformFormFields(tfFields []any) ([]*corev1.FormField, error) {
	fields := make([]*corev1.FormField, 0, len(tfFields))

	for idx, rawField := range tfFields {
		fieldMap, ok := rawField.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("field[%d] has invalid structure", idx)
		}

		field := &corev1.FormField{
			Id:   fieldMap["id"].(string),
			Name: fieldMap["name"].(string),
			Type: fieldMap["type"].(string),
		}

		rawConfig, hasConfig := fieldMap["config"]
		if hasConfig {
			configList := rawConfig.([]any)
			if len(configList) > 0 {
				configMap, ok := configList[0].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("field[%d].config has invalid structure", idx)
				}

				config, err := expandTerraformFormFieldConfig(configMap, idx)
				if err != nil {
					return nil, err
				}
				field.Config = config
			}
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func expandTerraformFormFieldConfig(configMap map[string]any, fieldIndex int) (*corev1.FormFieldConfig, error) {
	config := &corev1.FormFieldConfig{}

	rawOptions, hasOptions := configMap["option"]
	if hasOptions {
		optionRows := rawOptions.([]any)
		for optionIndex, row := range optionRows {
			optionMap, ok := row.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("field[%d].config.option[%d] has invalid structure", fieldIndex, optionIndex)
			}
			config.Options = append(config.Options, &corev1.FormFieldOption{
				Label: optionMap["label"].(string),
				Value: optionMap["value"].(string),
			})
		}
	}

	rawOptionsSource, hasOptionsSource := configMap["options_source"]
	if hasOptionsSource {
		optionsSourceList := rawOptionsSource.([]any)
		if len(optionsSourceList) > 0 {
			optionsSourceMap, ok := optionsSourceList[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("field[%d].config.options_source has invalid structure", fieldIndex)
			}

			commandRows := optionsSourceMap["command"].([]any)
			if len(commandRows) == 0 {
				return nil, fmt.Errorf("field[%d].config.options_source.command is required", fieldIndex)
			}
			commandMap, ok := commandRows[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("field[%d].config.options_source.command has invalid structure", fieldIndex)
			}

			optionsSource := &corev1.FormFieldOptionsSource{
				App: optionsSourceMap["app"].(string),
				Command: &corev1.FormFieldOptionsSourceCommand{
					Name: commandMap["name"].(string),
				},
				MachineUserId: optionsSourceMap["machine_user_id"].(string),
				Transform:     optionsSourceMap["transform"].(string),
			}

			rawInputJSON, _ := optionsSourceMap["input_json"].(string)
			rawInputJSON = strings.TrimSpace(rawInputJSON)
			hasInputJSON := rawInputJSON != ""

			rawInput := optionsSourceMap["input"]
			inputMap, hasInputValues, err := expandTerraformFormFieldOptionsSourceInput(rawInput)
			if err != nil {
				return nil, fmt.Errorf("field[%d].config.options_source.input has invalid structure", fieldIndex)
			}

			if hasInputJSON {
				if hasInputValues {
					return nil, fmt.Errorf("field[%d].config.options_source.input and input_json are mutually exclusive", fieldIndex)
				}

				inputJSONMap, err := parseJSONObjectString(rawInputJSON)
				if err != nil {
					return nil, fmt.Errorf("field[%d].config.options_source.input_json is invalid JSON: %w", fieldIndex, err)
				}

				inputStruct, err := structpb.NewStruct(inputJSONMap)
				if err != nil {
					return nil, fmt.Errorf("field[%d].config.options_source.input_json is invalid: %w", fieldIndex, err)
				}
				optionsSource.Input = inputStruct
			} else if hasInputValues {
				inputStruct, err := structpb.NewStruct(inputMap)
				if err != nil {
					return nil, fmt.Errorf("field[%d].config.options_source.input is invalid: %w", fieldIndex, err)
				}
				optionsSource.Input = inputStruct
			}

			config.OptionsSource = optionsSource
		}
	}

	if len(config.Options) == 0 && config.OptionsSource == nil {
		return nil, nil
	}

	return config, nil
}

func expandTerraformFormFieldOptionsSourceInput(rawInput any) (map[string]any, bool, error) {
	if rawInput == nil {
		return nil, false, nil
	}

	inputMap, ok := rawInput.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("input has invalid structure")
	}

	return inputMap, len(inputMap) > 0, nil
}

func normalizeProtoFormFields(protoFields []*corev1.FormField, currentFields []any) []any {
	tfFields := make([]any, 0, len(protoFields))
	currentFieldsByID := terraformFormFieldsByID(currentFields)

	for _, protoField := range protoFields {
		field := map[string]any{
			"id":   protoField.Id,
			"name": protoField.Name,
			"type": protoField.Type,
		}

		if protoField.Config != nil {
			config := normalizeProtoFormFieldConfig(protoField.Config, terraformFormFieldConfig(currentFieldsByID[protoField.Id]))
			if config != nil {
				field["config"] = []any{config}
			}
		}

		tfFields = append(tfFields, field)
	}

	return tfFields
}

func normalizeProtoFormFieldConfig(protoConfig *corev1.FormFieldConfig, currentConfig ...map[string]any) map[string]any {
	if protoConfig == nil {
		return nil
	}

	config := map[string]any{}

	if len(protoConfig.Options) > 0 {
		options := make([]any, 0, len(protoConfig.Options))
		for _, protoOption := range protoConfig.Options {
			options = append(options, map[string]any{
				"label": protoOption.Label,
				"value": protoOption.Value,
			})
		}
		config["option"] = options
	}

	if protoConfig.OptionsSource != nil {
		optionsSource := map[string]any{
			"app":             protoConfig.OptionsSource.App,
			"machine_user_id": protoConfig.OptionsSource.MachineUserId,
			"transform":       protoConfig.OptionsSource.Transform,
			"command": []any{
				map[string]any{
					"name": protoConfig.OptionsSource.Command.GetName(),
				},
			},
		}

		if protoConfig.OptionsSource.Input != nil {
			inputMap := protoConfig.OptionsSource.Input.AsMap()
			if terraformFormFieldConfigUsesInputJSON(currentConfig...) {
				optionsSource["input_json"] = mustCanonicalJSONString(inputMap)
			} else if terraformFormFieldConfigUsesInput(currentConfig...) {
				optionsSource["input"] = inputMap
			} else {
				optionsSource["input_json"] = mustCanonicalJSONString(inputMap)
			}
		}

		config["options_source"] = []any{optionsSource}
	}

	if len(config) == 0 {
		return nil
	}

	return config
}

func terraformFormFieldsByID(tfFields []any) map[string]map[string]any {
	fieldsByID := make(map[string]map[string]any, len(tfFields))
	for _, rawField := range tfFields {
		fieldMap, ok := rawField.(map[string]any)
		if !ok {
			continue
		}

		id, ok := fieldMap["id"].(string)
		if !ok || id == "" {
			continue
		}

		fieldsByID[id] = fieldMap
	}

	return fieldsByID
}

func terraformFormFieldConfig(tfField map[string]any) map[string]any {
	if tfField == nil {
		return nil
	}

	configList, ok := tfField["config"].([]any)
	if !ok || len(configList) == 0 {
		return nil
	}

	configMap, ok := configList[0].(map[string]any)
	if !ok {
		return nil
	}

	return configMap
}

func terraformFormFieldConfigUsesInputJSON(configs ...map[string]any) bool {
	if len(configs) == 0 || configs[0] == nil {
		return false
	}

	optionsSourceList, ok := configs[0]["options_source"].([]any)
	if !ok || len(optionsSourceList) == 0 {
		return false
	}

	optionsSourceMap, ok := optionsSourceList[0].(map[string]any)
	if !ok {
		return false
	}

	inputJSON, ok := optionsSourceMap["input_json"].(string)
	return ok && strings.TrimSpace(inputJSON) != ""
}

func terraformFormFieldConfigUsesInput(configs ...map[string]any) bool {
	if len(configs) == 0 || configs[0] == nil {
		return false
	}

	optionsSourceList, ok := configs[0]["options_source"].([]any)
	if !ok || len(optionsSourceList) == 0 {
		return false
	}

	optionsSourceMap, ok := optionsSourceList[0].(map[string]any)
	if !ok {
		return false
	}

	inputMap, ok := optionsSourceMap["input"].(map[string]any)
	return ok && len(inputMap) > 0
}

func validateJSONObjectString(value any, key string) ([]string, []error) {
	if _, err := parseJSONObjectString(value.(string)); err != nil {
		return nil, []error{fmt.Errorf("%s must be a valid JSON object: %w", key, err)}
	}

	return nil, nil
}

func canonicalizeJSONString(value any) string {
	canonicalJSON, err := parseJSONObjectString(value.(string))
	if err != nil {
		return value.(string)
	}

	return mustCanonicalJSONString(canonicalJSON)
}

func parseJSONObjectString(inputJSON string) (map[string]any, error) {
	var inputMap map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(inputJSON)), &inputMap); err != nil {
		return nil, err
	}

	if inputMap == nil {
		return nil, fmt.Errorf("expected JSON object")
	}

	return inputMap, nil
}

func mustCanonicalJSONString(inputMap map[string]any) string {
	canonicalJSON, err := json.Marshal(inputMap)
	if err != nil {
		return "{}"
	}

	return string(canonicalJSON)
}
