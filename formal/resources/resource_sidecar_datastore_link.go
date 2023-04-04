package resource

import (
	"context"
	"github.com/formalco/terraform-provider-formal/formal/apiv2"
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
		},
	}
}

func resourceSidecarDatastoreLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newSidecarDatastoreLink := apiv2.SidecarDatastoreLink{
		DatastoreId: d.Get("datastore_id").(string),
		SidecarId:   d.Get("sidecar_id").(string),
		Port:        d.Get("port").(int),
	}

	sidecarDatastoreLinkId, err := c.Grpc.CreateLink(ctx, newSidecarDatastoreLink)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(sidecarDatastoreLinkId)

	resourceSidecarDatastoreLinkRead(ctx, d, meta)
	return diags
}

func resourceSidecarDatastoreLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarDatastoreLinkId := d.Id()

	sidecarDatastoreLink, err := c.Grpc.GetLink(ctx, sidecarDatastoreLinkId)
	if err != nil {
		return diag.FromErr(err)
	}
	if sidecarDatastoreLink == nil {
		return diags
	}

	// Should map to all fields of
	d.Set("id", sidecarDatastoreLink.Id)
	d.Set("datastore_id", sidecarDatastoreLink.DatastoreId)
	d.Set("sidecar_id", sidecarDatastoreLink.SidecarId)
	d.Set("port", sidecarDatastoreLink.Port)

	d.SetId(sidecarDatastoreLink.Id)

	return diags
}

func resourceSidecarDatastoreLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarDatastoreLinkId := d.Id()

	err := c.Grpc.DeleteLink(ctx, sidecarDatastoreLinkId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
