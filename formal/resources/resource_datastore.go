package resource

import (
	"context"
	"fmt"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"
	"time"

	"github.com/formalco/terraform-provider-formal/formal/api"
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
				Description: "Technology of the Datastore: supported values are `snowflake`, `postgres`, and `redshift`.",
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
			"default_access_behavior": {
				// This description is used by the documentation generator and the language server.
				Description: "The default access behavior of the datastore. Accepted values are `allow` and `block`.",
				Type:        schema.TypeString,
				Required:    true,
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

	portInt, _ := d.Get("port").(int)

	newDatastore := api.DatastoreV2{
		Name:                    d.Get("name").(string),
		OriginalHostname:        d.Get("hostname").(string),
		Port:                    portInt,
		HealthCheckDbName:       d.Get("health_check_db_name").(string),
		Technology:              d.Get("technology").(string),
		DefaultAccessBehavior:   d.Get("default_access_behavior").(string),
		DbDiscoveryJobWaitTime:  d.Get("db_discovery_job_wait_time").(string),
		DbDiscoveryNativeRoleID: d.Get("db_discovery_native_role_id").(string),
	}

	datastoreId, err := c.Http.CreateDatastore(newDatastore)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "here:"+datastoreId)
	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(datastoreId)

	tflog.Info(ctx, "reading")
	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceDatastoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	tflog.Info(ctx, "reading.......")
	datastore, err := c.Http.GetDatastore(datastoreId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Datastore was deleted
			tflog.Warn(ctx, "The datastore was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", datastore.DsId)
	d.Set("name", datastore.Name)
	d.Set("hostname", datastore.OriginalHostname)
	d.Set("port", datastore.Port)
	d.Set("technology", datastore.Technology)
	d.Set("created_at", datastore.CreatedAt)
	d.Set("default_access_behavior", datastore.DefaultAccessBehavior)

	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(datastore.DsId)

	return diags
}

func resourceDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"name", "health_check_db_name", "default_access_behavior", "db_discovery_job_wait_time", "db_discovery_native_role_id"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		err := c.Http.UpdateDatastoreName(datastoreId, api.DatastoreV2{Name: name})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("health_check_db_name") {
		healthCheckName := d.Get("health_check_db_name").(string)
		err := c.Http.UpdateDatastoreHealthCheckDbName(datastoreId, api.DatastoreV2{HealthCheckDbName: healthCheckName})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("default_access_behavior") {
		defaultAccessBehavior := d.Get("default_access_behavior").(string)
		err := c.Http.UpdateDatastoreDefaultAcccessBehavior(datastoreId, api.DatastoreV2{DefaultAccessBehavior: defaultAccessBehavior})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("db_discovery_job_wait_time") || d.HasChange("db_discovery_native_role_id") {
		dbDiscoveryJobWaitTime := d.Get("db_discovery_job_wait_time").(string)
		dbDiscoveryNativeRoleID := d.Get("db_discovery_native_role_id").(string)
		err := c.Http.UpdateDatastoreDbDiscoveryConfig(datastoreId, api.DatastoreV2{DbDiscoveryJobWaitTime: dbDiscoveryJobWaitTime, DbDiscoveryNativeRoleID: dbDiscoveryNativeRoleID})
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

	err := c.Http.DeleteDatastore(dsId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
