package resource

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceConnectorListenerRule() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Connector Listener Rule with Formal.",
		CreateContext: resourceConnectorListenerRuleCreate,
		ReadContext:   resourceConnectorListenerRuleRead,
		UpdateContext: resourceConnectorListenerRuleUpdate,
		DeleteContext: resourceConnectorListenerRuleDelete,
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this connector listener.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_listener_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the listener this rule is associated with.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "The type of the rule. It can be either `any`, `resource` or `technology`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"any",
					"resource",
					"technology",
				}, false),
			},
			"rule": {
				// This description is used by the documentation generator and the language server.
				Description: "The rule to apply to the listener. It should be either the id of the resource or the name of the technology.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^(any|resource_.*|datastore_.*|postgres|mysql|snowflake|mongodb|redshift|mariadb|s3|dynamodb|documentdb|http|ssh|salesforce|kubernetes|clickhouse)$`),
					"Rule must start with 'resource_' or be a valid technology name (e.g., postgres, mysql, redis, mongodb) or 'any'",
				),
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this connector listener rule cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceConnectorListenerRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateConnectorListenerRuleRequest{
		ConnectorListenerId:   d.Get("connector_listener_id").(string),
		Type:                  d.Get("type").(string),
		Rule:                  d.Get("rule").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorListenerRule(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorListenerRule.Id)

	resourceConnectorListenerRuleRead(ctx, d, meta)

	return diags
}

func resourceConnectorListenerRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorListenerRuleId := d.Id()

	req := connect.NewRequest(&corev1.GetConnectorListenerRuleRequest{
		Id: connectorListenerRuleId,
	})

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorListenerRule(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector listener was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorListenerRule.Id)
	d.Set("connector_listener_id", res.Msg.ConnectorListenerRule.Listener.Id)
	d.Set("type", res.Msg.ConnectorListenerRule.Type)
	d.Set("rule", res.Msg.ConnectorListenerRule.Rule)
	d.Set("termination_protection", res.Msg.ConnectorListenerRule.TerminationProtection)

	d.SetId(res.Msg.ConnectorListenerRule.Id)

	return diags
}

func resourceConnectorListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorListenerRuleId := d.Id()

	fieldsThatCanChange := []string{"termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	terminationProtection := d.Get("termination_protection").(bool)

	req := connect.NewRequest(&corev1.UpdateConnectorListenerRuleRequest{
		Id:                    connectorListenerRuleId,
		TerminationProtection: &terminationProtection,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorListenerRule(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceConnectorListenerRuleRead(ctx, d, meta)

	return diags
}

func resourceConnectorListenerRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorListenerRuleId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Connector listener cannot be deleted because termination_protection is set to true")
	}

	req := connect.NewRequest(&corev1.DeleteConnectorListenerRuleRequest{
		Id: connectorListenerRuleId,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorListenerRule(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
