package resource

import (
	"context"
	"github.com/formalco/terraform-provider-formal/formal/apiv2"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDatastoreLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Linking a Datastore to a Sidecar in Formal.",

		CreateContext: resourceDatastoreLinkCreate,
		ReadContext:   resourceDatastoreLinkRead,
		// UpdateContext: resourceDatastoreLinkUpdate,
		DeleteContext: resourceDatastoreLinkDelete,

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

func resourceDatastoreLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newDatastoreLink := apiv2.SidecarDatastoreLink{
		DatastoreId: d.Get("datastore_id").(string),
		SidecarId:   d.Get("sidecar_id").(string),
		Port:        d.Get("port").(int),
	}

	datastoreLinkId, err := c.Grpc.CreateLink(ctx, newDatastoreLink)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(datastoreLinkId)

	resourceDatastoreLinkRead(ctx, d, meta)
	return diags
}

func resourceDatastoreLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreLinkId := d.Id()

	datastoreLink, err := c.Grpc.GetLink(ctx, datastoreLinkId)
	if err != nil {
		return diag.FromErr(err)
	}
	if datastoreLink == nil {
		return diags
	}

	// Should map to all fields of
	d.Set("id", datastoreLink.Id)
	d.Set("datastore_id", datastoreLink.DatastoreId)
	d.Set("sidecar_id", datastoreLink.SidecarId)
	d.Set("port", datastoreLink.Port)

	d.SetId(datastoreLink.Id)

	return diags
}

func resourceDatastoreLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	datastoreLinkId := d.Id()

	err := c.Grpc.DeleteLink(ctx, datastoreLinkId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
