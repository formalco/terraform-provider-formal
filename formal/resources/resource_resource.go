package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceResource() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Resource with Formal.",
		CreateContext: resourceDatastoreCreate,
		ReadContext:   resourceDatastoreRead,
		UpdateContext: resourceDatastoreUpdate,
		DeleteContext: resourceDatastoreDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the Resource.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Hostname of the Resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the Resource: supported values are `snowflake`, `postgres`, `redshift`, `mysql`, `mariadb`, `s3`, `dynamodb`, `mongodb`, `documentdb`, `http` and `ssh`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port your Resource is listening on.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Creation time of the Resource.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"environment": {
				// This description is used by the documentation generator and the language server.
				Description: "Environment for the Resource, options: DEV, TEST, QA, UAT, EI, PRE, STG, NON_PROD, PROD, CORP.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, the Resource cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	portInt, ok := d.Get("port").(int)
	if !ok {
		return diag.FromErr(fmt.Errorf("error reading port"))
	}

	Name := d.Get("name").(string)
	OriginalHostname := d.Get("hostname").(string)
	Port := portInt
	Technology := d.Get("technology").(string)
	Environment := d.Get("environment").(string)
	TerminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.SdkV2.ResourceServiceClient.CreateResource(ctx, connect.NewRequest(&corev1.CreateResourceRequest{
		Name:                  Name,
		Hostname:              OriginalHostname,
		Port:                  int32(Port),
		Technology:            Technology,
		Environment:           Environment,
		TerminationProtection: TerminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Resource.Id)

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	res, err := c.Grpc.SdkV2.ResourceServiceClient.GetResource(ctx, connect.NewRequest(&corev1.GetResourceRequest{Id: datastoreId}))
	if err != nil {
		return diag.FromErr(err)
	}
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Datastore was deleted
			tflog.Warn(ctx, "The datastore was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Resource.Id)
	d.Set("name", res.Msg.Resource.Name)
	d.Set("hostname", res.Msg.Resource.Hostname)
	d.Set("port", res.Msg.Resource.Port)
	d.Set("technology", res.Msg.Resource.Technology)
	d.Set("created_at", res.Msg.Resource.CreatedAt.AsTime().Unix())
	d.Set("environment", res.Msg.Resource.Environment)
	d.Set("termination_protection", res.Msg.Resource.TerminationProtection)

	d.SetId(res.Msg.Resource.Id)

	return diags
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"name", "health_check_db_name", "db_discovery_job_wait_time", "db_discovery_native_role_id", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	name := d.Get("name").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.SdkV2.ResourceServiceClient.UpdateResource(ctx, connect.NewRequest(&corev1.UpdateResourceRequest{Id: datastoreId, Name: &name, TerminationProtection: &terminationProtection}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dsId := d.Id()
	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Datastore cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.SdkV2.ResourceServiceClient.DeleteResource(ctx, connect.NewRequest(&corev1.DeleteResourceRequest{Id: dsId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
