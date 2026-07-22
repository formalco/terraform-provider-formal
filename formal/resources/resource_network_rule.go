package resource

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceNetworkRule() *schema.Resource {
	return &schema.Resource{
		Description: "Creating a Network Rule in Formal.",

		CreateContext: resourceNetworkRuleCreate,
		ReadContext:   resourceNetworkRuleRead,
		UpdateContext: resourceNetworkRuleUpdate,
		DeleteContext: resourceNetworkRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of this Network Rule.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Network Rule name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Network Rule description.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"cel_expression": {
				Description: "The CEL expression describing how this Network Rule matches and routes traffic.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "Defines the current status of the Network Rule. It can be one of the following: 'draft' or 'active'.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "active",
				ValidateFunc: validation.StringInSlice([]string{
					"draft",
					"active",
				}, false),
			},
			"termination_protection": {
				Description: "If set to true, this Network Rule cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"created_at": {
				Description: "When the Network Rule was created.",
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

func resourceNetworkRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	res, err := c.Grpc.Sdk.DesktopServiceClient.CreateDesktopRoutingRule(ctx, &corev1.CreateDesktopRoutingRuleRequest{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		CelExpression:         d.Get("cel_expression").(string),
		Status:                d.Get("status").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Rule.Id)

	return resourceNetworkRuleRead(ctx, d, meta)
}

func resourceNetworkRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	ruleID := d.Id()

	res, err := c.Grpc.Sdk.DesktopServiceClient.GetDesktopRoutingRule(ctx, &corev1.GetDesktopRoutingRuleRequest{Id: ruleID})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Network Rule with ID "+ruleID+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := setNetworkRuleState(d, res.Rule); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.DesktopServiceClient.UpdateDesktopRoutingRule(ctx, &corev1.UpdateDesktopRoutingRuleRequest{
		Id:                    d.Id(),
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		CelExpression:         d.Get("cel_expression").(string),
		Status:                d.Get("status").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkRuleRead(ctx, d, meta)
}

func resourceNetworkRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("Network Rule cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.DesktopServiceClient.DeleteDesktopRoutingRule(ctx, &corev1.DeleteDesktopRoutingRuleRequest{Id: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func setNetworkRuleState(d *schema.ResourceData, rule *corev1.DesktopRoutingRule) error {
	if err := d.Set("id", rule.Id); err != nil {
		return err
	}
	if err := d.Set("name", rule.Name); err != nil {
		return err
	}
	if err := d.Set("description", rule.Description); err != nil {
		return err
	}
	if err := d.Set("cel_expression", rule.CelExpression); err != nil {
		return err
	}
	if err := d.Set("status", rule.Status); err != nil {
		return err
	}
	if err := d.Set("termination_protection", rule.TerminationProtection); err != nil {
		return err
	}
	if rule.CreatedAt != nil {
		if err := d.Set("created_at", rule.CreatedAt.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	if rule.UpdatedAt != nil {
		if err := d.Set("updated_at", rule.UpdatedAt.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	d.SetId(rule.Id)
	return nil
}
