package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceTracker() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Tracker in Formal.",

		CreateContext: resourceTrackerCreate,
		ReadContext:   resourceTrackerRead,
		DeleteContext: resourceTrackerDelete,
		UpdateContext: resourceTrackerUpdate,
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
				ForceNew:    true,
			},
			"path": {
				// This description is used by the documentation generator and the language server.
				Description: "Path associated with this tracker.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
				ForceNew:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Tracker cannot be deleted.",
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

	// Maps to user-defined fields
	ResourceId := d.Get("resource_id").(string)
	Path := d.Get("path").(string)
	AllowClearTextValue := d.Get("allow_clear_text_value").(bool)

	res, err := c.Grpc.Sdk.TrackersServiceClient.CreateRowLevelTracker(ctx, connect.NewRequest(&corev1.CreateRowLevelTrackerRequest{
		ResourceId:          ResourceId,
		Path:                Path,
		AllowClearTextValue: AllowClearTextValue,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.RowLevelTracker.Id)

	resourceTrackerRead(ctx, d, meta)
	return diags
}

func resourceTrackerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	trackerId := d.Id()

	res, err := c.Grpc.Sdk.TrackersServiceClient.GetRowLevelTracker(ctx, connect.NewRequest(&corev1.GetRowLevelTrackerRequest{Id: trackerId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Tracker was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.RowLevelTacker.Id)
	d.Set("resource_id", res.Msg.RowLevelTacker.ResourceId)
	d.Set("path", res.Msg.RowLevelTacker.Path)
	d.Set("created_at", res.Msg.RowLevelTacker.CreatedAt.AsTime().Unix())
	d.Set("allow_clear_text_value", res.Msg.RowLevelTacker.AllowClearTextValue)

	d.SetId(res.Msg.RowLevelTacker.Id)

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

	_, err := c.Grpc.Sdk.TrackersServiceClient.DeleteRowLevelTracker(ctx, connect.NewRequest(&corev1.DeleteRowLevelTrackerRequest{Id: trackerId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceTrackerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Row Level Trackers are immutable. Please create a new tracker.")
}
