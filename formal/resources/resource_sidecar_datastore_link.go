package resource

import (
	"context"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSidecarDatastoreLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Linking a Datastore to a Sidecar in Formal.",

		CreateContext: resourceSidecarDatastoreLinkCreate,
		ReadContext:   resourceSidecarDatastoreLinkRead,
		// UpdateContext: resourceDatastoreLinkUpdate,
		DeleteContext: resourceSidecarDatastoreLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Datastore ID to be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"sidecar_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Sidecar ID that should be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "Port.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Sidecar Datastore Link cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceSidecarDatastoreLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.SidecarServiceClient.CreateSidecarDatastoreLink(ctx, connect.NewRequest(&adminv1.CreateSidecarDatastoreLinkRequest{
		DatastoreId:           d.Get("datastore_id").(string),
		SidecarId:             d.Get("sidecar_id").(string),
		Port:                  int32(d.Get("port").(int)),
		TerminationProtection: d.Get("termination_protection").(bool),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.LinkId)

	resourceSidecarDatastoreLinkRead(ctx, d, meta)
	return diags
}

func resourceSidecarDatastoreLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecarDatastoreLink(ctx, connect.NewRequest(&adminv1.UpdateSidecarDatastoreLinkRequest{
			Id:                    d.Id(),
			TerminationProtection: &terminationProtection,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceSidecarDatastoreLinkRead(ctx, d, meta)

	return diags
}

func resourceSidecarDatastoreLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarDatastoreLinkId := d.Id()

	res, err := c.Grpc.Sdk.SidecarServiceClient.GetLinkById(ctx, connect.NewRequest(&adminv1.GetLinkByIdRequest{Id: sidecarDatastoreLinkId}))
	if err != nil {
		return diag.FromErr(err)
	}

	// Should map to all fields of
	d.Set("id", res.Msg.Link.Id)
	d.Set("datastore_id", res.Msg.Link.DatastoreId)
	d.Set("sidecar_id", res.Msg.Link.SidecarId)
	d.Set("port", res.Msg.Link.Port)
	d.Set("termination_protection", res.Msg.Link.TerminationProtection)

	d.SetId(res.Msg.Link.Id)

	return diags
}

func resourceSidecarDatastoreLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarDatastoreLinkId := d.Id()

	_, err := c.Grpc.Sdk.SidecarServiceClient.DeleteSidecarDatastoreLink(ctx, connect.NewRequest(&adminv1.DeleteSidecarDatastoreLinkRequest{LinkId: sidecarDatastoreLinkId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
