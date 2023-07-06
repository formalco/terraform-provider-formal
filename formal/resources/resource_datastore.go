package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDatastore() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Datastore with Formal.",
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
				Description: "The ID of the Datastore.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the Datastore.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Hostname of the Datastore.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the Datastore: supported values are `snowflake`, `postgres`, `redshift`, `mysql`, `mariadb`, `s3`, `http` and `ssh`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"health_check_db_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Database name to use for health checks. Default `postgres`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port your Datastore is listening on.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Creation time of the datastore.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"db_discovery_native_role_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The `native_role_id` of the Native Role to be used for the discovery job.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"db_discovery_job_wait_time": {
				// This description is used by the documentation generator and the language server.
				Description: "The wait time for the discovery job.",
				Type:        schema.TypeString,
				Optional:    true,
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

	Name := d.Get("name").(string)
	OriginalHostname := d.Get("hostname").(string)
	Port := portInt
	HealthCheckDbName := d.Get("health_check_db_name").(string)
	Technology := d.Get("technology").(string)
	DbDiscoveryJobWaitTime := d.Get("db_discovery_job_wait_time").(string)
	DbDiscoveryNativeRoleID := d.Get("db_discovery_native_role_id").(string)

	res, err := c.Grpc.Sdk.DataStoreServiceClient.CreateDatastore(ctx, connect.NewRequest(&adminv1.CreateDatastoreRequest{
		Name:                    Name,
		Hostname:                OriginalHostname,
		Port:                    int32(Port),
		Technology:              Technology,
		HealthCheckDbName:       HealthCheckDbName,
		DbDiscoveryJobWaitTime:  DbDiscoveryJobWaitTime,
		DbDiscoveryNativeRoleId: DbDiscoveryNativeRoleID,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(res.Msg.Id)

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceDatastoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	res, err := c.Grpc.Sdk.DataStoreServiceClient.GetDatastore(ctx, connect.NewRequest(&adminv1.GetDatastoreRequest{Id: datastoreId}))
	if err != nil {
		return diag.FromErr(err)
	}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Datastore was deleted
			tflog.Warn(ctx, "The datastore was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Datastore.DatastoreId)
	d.Set("name", res.Msg.Datastore.Name)
	d.Set("hostname", res.Msg.Datastore.Hostname)
	d.Set("port", res.Msg.Datastore.Port)
	d.Set("technology", res.Msg.Datastore.Technology)
	d.Set("created_at", res.Msg.Datastore.CreatedAt)

	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(res.Msg.Datastore.DatastoreId)

	return diags
}

func resourceDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"name", "health_check_db_name", "db_discovery_job_wait_time", "db_discovery_native_role_id"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, err := c.Grpc.Sdk.DataStoreServiceClient.UpdateDatastoreName(ctx, connect.NewRequest(&adminv1.UpdateDatastoreNameRequest{Id: datastoreId, Name: name}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("health_check_db_name") {
		healthCheckName := d.Get("health_check_db_name").(string)
		_, err := c.Grpc.Sdk.DataStoreServiceClient.UpdateDataStoreHealthCheckDbName(ctx, connect.NewRequest(&adminv1.UpdateDataStoreHealthCheckDbNameRequest{Id: datastoreId, HealthCheckDbName: healthCheckName}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("db_discovery_job_wait_time") || d.HasChange("db_discovery_native_role_id") {
		dbDiscoveryJobWaitTime := d.Get("db_discovery_job_wait_time").(string)
		dbDiscoveryNativeRoleID := d.Get("db_discovery_native_role_id").(string)
		_, err := c.Grpc.Sdk.DataStoreServiceClient.UpdateDbDiscoveryConfig(ctx, connect.NewRequest(&adminv1.UpdateDbDiscoveryConfigRequest{Id: datastoreId, DbDiscoveryJobWaitTime: dbDiscoveryJobWaitTime, DbDiscoveryNativeRoleId: dbDiscoveryNativeRoleID}))
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

	_, err := c.Grpc.Sdk.DataStoreServiceClient.DeleteDatastore(ctx, connect.NewRequest(&adminv1.DeleteDatastoreRequest{Id: dsId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
