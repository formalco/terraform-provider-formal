package resource

import (
	"context"
	"fmt"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/protobuf/types/known/structpb"

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

func resourceFormCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	fields, err := expandTerraformFormFields(d.Get("field").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.WorkflowServiceClient.CreateForm(ctx, connect.NewRequest(&corev1.CreateFormRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Fields:      fields,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Form.Id)

	return resourceFormRead(ctx, d, meta)
}

func resourceFormRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	formID := d.Id()

	res, err := c.Grpc.Sdk.WorkflowServiceClient.GetForm(ctx, connect.NewRequest(&corev1.GetFormRequest{Id: formID}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Form with ID "+formID+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	form := res.Msg.Form

	if err := d.Set("id", form.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", form.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", form.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("field", flattenProtoFormFields(form.Fields)); err != nil {
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

func resourceFormUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("field") {
		fields, err := expandTerraformFormFields(d.Get("field").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = c.Grpc.Sdk.WorkflowServiceClient.UpdateForm(ctx, connect.NewRequest(&corev1.UpdateFormRequest{
			Id:          d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Fields:      fields,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFormRead(ctx, d, meta)
}

func resourceFormDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.WorkflowServiceClient.DeleteForm(ctx, connect.NewRequest(&corev1.DeleteFormRequest{Id: d.Id()}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandTerraformFormFields(tfFields []interface{}) ([]*corev1.FormField, error) {
	fields := make([]*corev1.FormField, 0, len(tfFields))

	for idx, rawField := range tfFields {
		fieldMap, ok := rawField.(map[string]interface{})
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
			configList := rawConfig.([]interface{})
			if len(configList) > 0 {
				configMap, ok := configList[0].(map[string]interface{})
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

func expandTerraformFormFieldConfig(configMap map[string]interface{}, fieldIndex int) (*corev1.FormFieldConfig, error) {
	config := &corev1.FormFieldConfig{}

	rawOptions, hasOptions := configMap["option"]
	if hasOptions {
		optionRows := rawOptions.([]interface{})
		for optionIndex, row := range optionRows {
			optionMap, ok := row.(map[string]interface{})
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
		optionsSourceList := rawOptionsSource.([]interface{})
		if len(optionsSourceList) > 0 {
			optionsSourceMap, ok := optionsSourceList[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("field[%d].config.options_source has invalid structure", fieldIndex)
			}

			commandRows := optionsSourceMap["command"].([]interface{})
			if len(commandRows) == 0 {
				return nil, fmt.Errorf("field[%d].config.options_source.command is required", fieldIndex)
			}
			commandMap, ok := commandRows[0].(map[string]interface{})
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

			rawInput, hasInput := optionsSourceMap["input"]
			if hasInput && rawInput != nil {
				inputMap, ok := rawInput.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("field[%d].config.options_source.input has invalid structure", fieldIndex)
				}

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

func flattenProtoFormFields(protoFields []*corev1.FormField) []interface{} {
	tfFields := make([]interface{}, 0, len(protoFields))

	for _, protoField := range protoFields {
		field := map[string]interface{}{
			"id":   protoField.Id,
			"name": protoField.Name,
			"type": protoField.Type,
		}

		if protoField.Config != nil {
			config := flattenProtoFormFieldConfig(protoField.Config)
			if config != nil {
				field["config"] = []interface{}{config}
			}
		}

		tfFields = append(tfFields, field)
	}

	return tfFields
}

func flattenProtoFormFieldConfig(protoConfig *corev1.FormFieldConfig) map[string]interface{} {
	if protoConfig == nil {
		return nil
	}

	config := map[string]interface{}{}

	if len(protoConfig.Options) > 0 {
		options := make([]interface{}, 0, len(protoConfig.Options))
		for _, protoOption := range protoConfig.Options {
			options = append(options, map[string]interface{}{
				"label": protoOption.Label,
				"value": protoOption.Value,
			})
		}
		config["option"] = options
	}

	if protoConfig.OptionsSource != nil {
		optionsSource := map[string]interface{}{
			"app":             protoConfig.OptionsSource.App,
			"machine_user_id": protoConfig.OptionsSource.MachineUserId,
			"transform":       protoConfig.OptionsSource.Transform,
			"command": []interface{}{
				map[string]interface{}{
					"name": protoConfig.OptionsSource.Command.GetName(),
				},
			},
		}

		if protoConfig.OptionsSource.Input != nil {
			optionsSource["input"] = protoConfig.OptionsSource.Input.AsMap()
		}

		config["options_source"] = []interface{}{optionsSource}
	}

	if len(config) == 0 {
		return nil
	}

	return config
}
