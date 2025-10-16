package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
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
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the Resource: supported values are `snowflake`, `postgres`, `redshift`, `mysql`, `mariadb`, `s3`, `dynamodb`, `mongodb`, `documentdb`, `http`, `clickhouse`, `redis`, `web` and `ssh`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"technology_provider": {
				// This description is used by the documentation generator and the language server.
				Description: "For SSH resources, if the backend connection is SSM, supported values are `aws-ec2`, and `aws-ecs`",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port your Resource is listening on.",
				Type:        schema.TypeInt,
				Required:    true,
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
				Deprecated:  "This field is deprecated and will be removed in a future release.",
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, the Resource cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"space_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Space to create the Resource in.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
		},
	}
}

func resourceDatastoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	portInt, ok := d.Get("port").(int)
	if !ok {
		return diag.FromErr(fmt.Errorf("error reading port"))
	}

	name := d.Get("name").(string)
	hostname := d.Get("hostname").(string)
	port := portInt
	technology := d.Get("technology").(string)
	environment := d.Get("environment").(string)
	terminationProtection := d.Get("termination_protection").(bool)
	spaceId := d.Get("space_id").(string)

	msg := &corev1.CreateResourceRequest{
		Name:                  name,
		Hostname:              hostname,
		Port:                  int32(port),
		Technology:            technology,
		Environment:           environment,
		TerminationProtection: terminationProtection,
	}

	if spaceId != "" {
		msg.SpaceId = &spaceId
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResource(ctx, connect.NewRequest(msg))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Resource.Id)

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceDatastoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResource(ctx, connect.NewRequest(&corev1.GetResourceRequest{Id: datastoreId}))
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
	d.Set("technology_provider", res.Msg.Resource.Provider)
	d.Set("environment", res.Msg.Resource.Environment)
	d.Set("termination_protection", res.Msg.Resource.TerminationProtection)
	if res.Msg.Resource.Space != nil {
		d.Set("space_id", res.Msg.Resource.Space.Id)
	}
	d.SetId(res.Msg.Resource.Id)

	return diags
}

func resourceDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"name", "environment", "hostname", "termination_protection", "space_id"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResource(ctx, connect.NewRequest(&corev1.UpdateResourceRequest{
			Id:   datastoreId,
			Name: &name,
		}))
		if err != nil {
			return diag.FromErr(err)
		}

	}

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResource(ctx, connect.NewRequest(&corev1.UpdateResourceRequest{
			Id:                    datastoreId,
			TerminationProtection: &terminationProtection,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("space_id") {
		spaceId := d.Get("space_id").(string)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResource(ctx, connect.NewRequest(&corev1.UpdateResourceRequest{
			Id:      datastoreId,
			SpaceId: &spaceId,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("hostname") {
		hostname := d.Get("hostname").(string)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResource(ctx, connect.NewRequest(&corev1.UpdateResourceRequest{
			Id:       datastoreId,
			Hostname: &hostname,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceDatastoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dsId := d.Id()
	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Datastore cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResource(ctx, connect.NewRequest(&corev1.DeleteResourceRequest{Id: dsId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
