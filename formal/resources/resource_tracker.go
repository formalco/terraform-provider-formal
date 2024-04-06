package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceTracker() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Policy in Formal.",

		CreateContext: resourceTrackerCreate,
		DeleteContext: resourceTrackerDelete,
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
				Description: "Tracker linked to the following resource id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"path": {
				// This description is used by the documentation generator and the language server.
				Description: "Path associated with this tracker.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of this Tracker.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the policy was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"allow_clear_text_value": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Tracker allow clear text value.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceTrackerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var owners []*corev1.Owner
	for _, owner := range d.Get("owners").([]*corev1.Owner) {
		owners = append(owners, owner)
	}

	// Maps to user-defined fields
	ResourceId := d.Get("resource_id").(string)
	Path := d.Get("path").(string)
	AllowClearTextValue := d.Get("allow_clear_text_value").(bool)

	res, err := c.Grpc.Sdk.RowLevelTrackerServiceClient.CreateRowLevelTracker(ctx, connect.NewRequest(&corev1.CreateRowLevelTrackerRequest{
		ResourceId:          ResourceId,
		Path:                Path,
		AllowClearTextValue: AllowClearTextValue,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.RowLevelTracker.Id)

	resourcePolicyRead(ctx, d, meta)
	return diags
}

func resourceTrackerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	trackerId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Policy cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.RowLevelTrackerServiceClient.DeleteRowLevelTracker(ctx, connect.NewRequest(&corev1.DeleteRowLevelTrackerRequest{Id: trackerId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
