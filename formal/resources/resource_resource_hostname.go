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

func ResourceResourceHostname() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Resource Hostname with Formal.",
		CreateContext: resourceResourceHostnameCreate,
		ReadContext:   resourceResourceHostnameRead,
		UpdateContext: resourceResourceHostnameUpdate,
		DeleteContext: resourceResourceHostnameDelete,
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
				Description: "The ID of this Resource Hostname.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Resource this hostname is linked to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "The name of this Resource Hostname.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The hostname for this Resource hostname.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this resource hostname cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceResourceHostnameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateResourceHostnameRequest{
		ResourceId:            d.Get("resource_id").(string),
		Name:                  d.Get("name").(string),
		Hostname:              d.Get("hostname").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceHostname(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ResourceHostname.Id)
	resourceResourceHostnameRead(ctx, d, meta)
	return diags
}

func resourceResourceHostnameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceHostnameId := d.Id()

	req := connect.NewRequest(&corev1.GetResourceHostnameRequest{
		Id: resourceHostnameId,
	})

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceHostname(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Resource Hostname was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ResourceHostname.Id)
	d.Set("resource_id", res.Msg.ResourceHostname.Resource.Id)
	d.Set("hostname", res.Msg.ResourceHostname.Hostname)
	d.Set("name", res.Msg.ResourceHostname.Name)

	d.SetId(res.Msg.ResourceHostname.Id)

	return diags
}

func resourceResourceHostnameUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	resourceHostnameId := d.Id()

	fieldsThatCanChange := []string{"hostname", "name", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	hostname := d.Get("hostname").(string)
	name := d.Get("name").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	req := connect.NewRequest(&corev1.UpdateResourceHostnameRequest{
		Id:                    resourceHostnameId,
		Name:                  &name,
		Hostname:              &hostname,
		TerminationProtection: &terminationProtection,
	})

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceHostname(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceResourceHostnameRead(ctx, d, meta)
	return diags
}

func resourceResourceHostnameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	resourceHostnameId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Resource Hostname cannot be deleted because termination_protection is set to true")
	}

	req := connect.NewRequest(&corev1.DeleteResourceHostnameRequest{
		Id: resourceHostnameId,
	})

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceHostname(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
