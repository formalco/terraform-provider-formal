package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSidecarListenerRule() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Sidecar Listener Rule with Formal.",
		CreateContext: resourceSidecarListenerRuleCreate,
		ReadContext:   resourceSidecarListenerRuleRead,
		UpdateContext: resourceSidecarListenerRuleUpdate,
		DeleteContext: resourceSidecarListenerRuleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the listener rule.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sidecar_listener_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the sidecar listener.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "The type of the rule to apply to the sidecar listener.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"rule": {
				// This description is used by the documentation generator and the language server.
				Description: "The rule to apply to the sidecar listener.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceSidecarListenerRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarListenerId := d.Get("sidecar_listener_id").(string)
	ruleType := d.Get("type").(string)
	rule := d.Get("rule").(string)

	res, err := c.Grpc.Sdk.ListenersServiceClient.CreateSidecarListenerRule(ctx, connect.NewRequest(&corev1.CreateSidecarListenerRuleRequest{
		SidecarListenerId: sidecarListenerId,
		Type:              ruleType,
		Rule:              rule,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)

	resourceSidecarListenerRuleRead(ctx, d, meta)

	return diags
}

func resourceSidecarListenerRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarListenerRuleId := d.Id()

	res, err := c.Grpc.Sdk.ListenersServiceClient.GetSidecarListenerRule(ctx, connect.NewRequest(&corev1.GetSidecarListenerRuleRequest{Id: sidecarListenerRuleId}))
	if err != nil {
		return diag.FromErr(err)
	}
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Sidecar listener rule was deleted
			tflog.Warn(ctx, "The sidecar listener was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Id)
	d.Set("sidecar_listener_id", res.Msg.Type)
	d.Set("type", res.Msg.Type)
	d.Set("rule", res.Msg.Rule)
	d.SetId(res.Msg.Id)

	return diags
}

func resourceSidecarListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarListenerRuleId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"sidecar_listener_id", "type", "rule"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	sidecarListenerId := d.Get("sidecar_listener_id").(string)
	ruleType := d.Get("type").(string)
	rule := d.Get("rule").(string)

	req := connect.NewRequest(&corev1.UpdateSidecarListenerRuleRequest{
		Id:                sidecarListenerRuleId,
		SidecarListenerId: sidecarListenerId,
		Type:              ruleType,
		Rule:              rule,
	})
	_, err := c.Grpc.Sdk.ListenersServiceClient.UpdateSidecarListenerRule(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSidecarListenerRuleRead(ctx, d, meta)

	return diags
}

func resourceSidecarListenerRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarListenerRuleId := d.Id()
	req := connect.NewRequest(&corev1.DeleteSidecarListenerRuleRequest{
		Id: sidecarListenerRuleId,
	})

	_, err := c.Grpc.Sdk.ListenersServiceClient.DeleteSidecarListenerRule(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
