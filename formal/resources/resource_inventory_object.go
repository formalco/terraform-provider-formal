package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceInventoryObject() *schema.Resource {
	return &schema.Resource{
		Description:   "Registering an inventory object (db, schema, table, column, sub-column) with Formal. Useful for seeding the inventory in test fixtures so that connectors load it at startup instead of relying on inline discovery.",
		CreateContext: resourceInventoryObjectCreate,
		ReadContext:   resourceInventoryObjectRead,
		DeleteContext: resourceInventoryObjectDelete,
		SchemaVersion: 1,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the inventory object.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "Resource (datastore) ID this object belongs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description: "Object type. One of `db`, `schema`, `table`, `column`, `sub-column`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"db", "schema", "table", "column", "sub-column",
				}, false),
			},
			"path": {
				Description: "Hierarchical path of the object (e.g. `mydb.public.users.email`).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Last segment of the path (e.g. `email`).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"data_type": {
				Description: "Column data type (e.g. `varchar`). Required when `type` is `column`, ignored otherwise.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"sub_type": {
				Description: "Sub-column type. One of `json`, `hstore`. Required when `type` is `sub-column`, ignored otherwise.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"json", "hstore",
				}, false),
			},
		},
	}
}

func resourceInventoryObjectCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	objectType := d.Get("type").(string)
	path := d.Get("path").(string)
	name := d.Get("name").(string)

	req := &corev1.CreateInventoryObjectRequest{
		DatastoreId: d.Get("resource_id").(string),
		ObjectType:  objectType,
	}

	switch objectType {
	case "db":
		req.Db = &corev1.CreateInventoryObjectRequest_Db{Path: path, Name: name}
	case "schema":
		req.Schema = &corev1.CreateInventoryObjectRequest_Schema{Path: path, Name: name}
	case "table":
		req.Table = &corev1.CreateInventoryObjectRequest_Table{Path: path, Name: name}
	case "column":
		dataType := d.Get("data_type").(string)
		if dataType == "" {
			return diag.Errorf("`data_type` is required when `type` is `column`")
		}
		req.Column = &corev1.CreateInventoryObjectRequest_Column{Path: path, Name: name, DataType: dataType}
	case "sub-column":
		subType := d.Get("sub_type").(string)
		if subType == "" {
			return diag.Errorf("`sub_type` is required when `type` is `sub-column`")
		}
		req.SubColumn = &corev1.CreateInventoryObjectRequest_SubColumn{Path: path, Name: name, SubType: subType}
	}

	res, err := c.Grpc.Sdk.InventoryServiceClient.CreateInventoryObject(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)
	return resourceInventoryObjectRead(ctx, d, meta)
}

func resourceInventoryObjectRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	res, err := c.Grpc.Sdk.InventoryServiceClient.GetInventoryObject(ctx, connect.NewRequest(&corev1.GetInventoryObjectRequest{Id: d.Id()}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The inventory object was not found; it may have been deleted outside Terraform.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	switch obj := res.Msg.Object.GetObject().(type) {
	case *corev1.InventoryObject_Db:
		d.Set("resource_id", obj.Db.ResourceId)
		d.Set("type", "db")
		d.Set("path", obj.Db.Path)
		d.Set("name", obj.Db.Name)
	case *corev1.InventoryObject_Schema:
		d.Set("resource_id", obj.Schema.ResourceId)
		d.Set("type", "schema")
		d.Set("path", obj.Schema.Path)
		d.Set("name", obj.Schema.Name)
	case *corev1.InventoryObject_Table:
		d.Set("resource_id", obj.Table.ResourceId)
		d.Set("type", "table")
		d.Set("path", obj.Table.Path)
		d.Set("name", obj.Table.Name)
	case *corev1.InventoryObject_Column:
		d.Set("resource_id", obj.Column.ResourceId)
		d.Set("type", "column")
		d.Set("path", obj.Column.Path)
		d.Set("name", obj.Column.Name)
		d.Set("data_type", obj.Column.DataType)
	case *corev1.InventoryObject_SubColumn:
		// The API does not echo sub_type back, so it is left as configured.
		d.Set("resource_id", obj.SubColumn.ResourceId)
		d.Set("type", "sub-column")
		d.Set("path", obj.SubColumn.Path)
		d.Set("name", obj.SubColumn.Name)
	}
	return nil
}

func resourceInventoryObjectDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.InventoryServiceClient.DeleteInventoryObject(ctx, connect.NewRequest(&corev1.DeleteInventoryObjectRequest{Id: d.Id()}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
