package resource

import (
	"context"
	"fmt"
	"strings"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceHealthCheck() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Policy in Formal.",

		CreateContext: resourceHealthCheckCreate,
		ReadContext:   resourceHealthCheckRead,
		DeleteContext: resourceHealthCheckDelete,
		UpdateContext: resourceHealthCheckUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePolicyInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePolicyStateUpgradeV0,
			},
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID linked to the following health check.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"database_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Database associated with the health check.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the Health Check.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the policy was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceHealthCheckCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	ResourceId := d.Get("resource_id").(string)
	DatabaseName := d.Get("database_name").(string)

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceHealthCheck(ctx, connect.NewRequest(&corev1.CreateResourceHealthCheckRequest{
		ResourceId:   ResourceId,
		DatabaseName: DatabaseName,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ResourceHealthCheck.Id)

	resourcePolicyRead(ctx, d, meta)
	return diags
}

func resourceHealthCheckRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := corev1.GetResourceHealthCheckRequest_ResourceHealthCheckId{
		ResourceHealthCheckId: d.Id(),
	}

	_, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceHealthCheck(ctx, connect.NewRequest(&corev1.GetResourceHealthCheckRequest{Id: &id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceHealthCheckUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	resourceHealthCheckId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"database_name"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	DatabaseName := d.Get("database_name").(string)

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceHealthCheck(ctx, connect.NewRequest(&corev1.UpdateResourceHealthCheckRequest{Id: resourceHealthCheckId, DatabaseName: &DatabaseName}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceHealthCheckRead(ctx, d, meta)

	return diags
}

func resourceHealthCheckDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceHealthCheckId := d.Id()

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceHealthCheck(ctx, connect.NewRequest(&corev1.DeleteResourceHealthCheckRequest{Id: resourceHealthCheckId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
