package resource

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this datastore.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Hostname of the datastore.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the datastore: supported values are `snowflake`, `postgres`, and `redshift`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"deployment_type": {
				// This description is used by the documentation generator and the language server.
				Description: "How the sidecar for this datastore should be deployed: `saas`, `managed`, or `onprem`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fail_open": {
				// This description is used by the documentation generator and the language server.
				Description: "Configure DNS failover from the sidecar to the original datastore. In the unlikely case where the sidecar is healthy, having this value of `true` will forward traffic to the original database. Default `false`.",
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
			},
			"network_type": {
				// This description is used by the documentation generator and the language server.
				Description: "Configure the sidecar network type. Value can be `internet-facing`, `internal` or `internet-and-internal`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"username": {
				// This description is used by the documentation generator and the language server.
				Description: "Username for the original datastore that the sidecar should use. Please be sure to set this secret via Terraform environment variables.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"password": {
				// This description is used by the documentation generator and the language server.
				Description: "Password for the original datastore that the sidecar should use. Please be sure to set this secret via Terraform environment variables.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"health_check_db_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Database name to use for health checks. Default `postgres`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cloud_provider": {
				// This description is used by the documentation generator and the language server.
				Description: "Cloud Provider that the sidecar sholud deploy in. Supported values at the moment are `aws`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port your datastore is listening on. Required if your `technology` is `postgres` or `redshift`.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "The cloud region the sidecar should be deployed in. For SaaS deployment models, supported values are `eu-west-1`, `eu-west-3`, `us-east-1`, and `us-west-2`",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"cloud_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Required for managed cloud - the Formal ID for the connected Cloud Account. You can find this after creating the connection in the Formal Console.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			// "id": {
			// 	// This description is used by the documentation generator and the language server.
			// 	Description: "Formal ID for this record.",
			// 	Type:        schema.TypeString,
			// 	Computed:    true,
			// },
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID for the datastore.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"org_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for your organisation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"stack_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Name of the CloudFormation stack if deployed as managed.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"formal_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The hostname of the created sidcar.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"net_stack_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Net Stack ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Creation time of the sidecar.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"global_kms_decrypt": {
				// This description is used by the documentation generator and the language server.
				Description: "Enable all Field Encryptions created by this sidecar to be decrypted by other sidecars.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"dataplane_id": {
				// This description is used by the documentation generator and the language server.
				Description: "If deployment_type is managed, this is the ID of the Dataplane",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"formal_control_plane_tls_certificate": {
				// This description is used by the documentation generator and the language server.
				Description: "If deployment_type is onprem, this is the Control Plane TLS Certificate to add to the deployed Sidecar.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceDatastoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*(api.Client))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	portInt, _ := d.Get("port").(int)

	newDatastore := api.DataStoreInfra{
		// Id
		// DsId
		// OrgId
		// StackName
		Name:              d.Get("name").(string),
		Hostname:          d.Get("hostname").(string),
		Port:              portInt,
		Username:          d.Get("username").(string),
		Password:          d.Get("password").(string),
		HealthCheckDbName: d.Get("health_check_db_name").(string),
		// FormalHostName
		Technology:     d.Get("technology").(string),
		CloudProvider:  d.Get("cloud_provider").(string),
		CloudRegion:    d.Get("cloud_region").(string),
		DeploymentType: d.Get("deployment_type").(string),
		CloudAccountID: d.Get("cloud_account_id").(string),
		// NetStackId:
		FailOpen:    d.Get("fail_open").(bool),
		NetworkType: d.Get("network_type").(string),
		DataplaneID: d.Get("dataplane_id").(string),
		// CreateAt
	}

	res, err := client.CreateDatastore(newDatastore)
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdDatastore, err := client.GetDatastoreForStatus(res.DsId)
		if err != nil {
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors+1)+" checking on DatastoreStatus: ", map[string]interface{}{"err": err})
				currentErrors += 1
				time.Sleep(15 * time.Second)
				continue
			}
		}

		if createdDatastore == nil {
			err = errors.New("datastore with the given ID not found. It may have been deleted")
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "Deployed is "+fmt.Sprint(createdDatastore.Deployed))
		// Check status
		if createdDatastore.Deployed {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	fullKMSDecryption := d.Get("global_kms_decrypt").(bool)
	if fullKMSDecryption {
		client.UpdateDatastoreGlobalKMSEncrypt(res.DsId, api.DataStoreInfra{FullKMSDecryption: true})
	}

	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(res.DsId)

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceDatastoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	datastoreId := d.Id()

	datastore, err := client.GetDatastore(datastoreId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Datastore was deleted
			tflog.Warn(ctx, "The datastore was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// d.Set("id", datastore.Id)
	d.Set("datastore_id", datastore.DsId)
	d.Set("org_id", datastore.OrganisationID)
	d.Set("stack_name", datastore.StackName)
	d.Set("name", datastore.Name)
	d.Set("hostname", datastore.Hostname)
	d.Set("port", datastore.Port)
	d.Set("username", datastore.Username)
	d.Set("password", datastore.Password)
	d.Set("formal_hostname", datastore.FormalHostname)
	d.Set("technology", datastore.Technology)
	d.Set("cloud_provider", datastore.CloudProvider)
	d.Set("cloud_region", datastore.CloudRegion)
	d.Set("deployment_type", datastore.DeploymentType)
	d.Set("cloud_account_id", datastore.CloudAccountID)
	d.Set("net_stack_id", datastore.NetStackId)
	d.Set("fail_open", datastore.FailOpen)
	d.Set("network_type", datastore.NetworkType)
	d.Set("created_at", datastore.CreatedAt)
	d.Set("global_kms_decrypt", datastore.FullKMSDecryption)
	d.Set("dataplane_id", datastore.DataplaneID)

	if datastore.DeploymentType == "onprem" {
		tlsCert, err := client.GetDatastoreTlsCert(datastoreId)
		if err != nil {
			return diag.FromErr(err)
		}
		if *tlsCert == "" {
			return diag.Errorf("The TLS Certificate was not found. Please contact the Formal team for support.")
		}

		d.Set("formal_control_plane_tls_certificate", *tlsCert)
	}

	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(datastore.DsId)

	return diags
}

func resourceDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	var diags diag.Diagnostics

	datastoreId := d.Id()

	// Only enable updates to these fields, err otherwise
	if d.HasChange("global_kms_decrypt") || d.HasChange("username") || d.HasChange("password") || d.HasChange("name") || d.HasChange("health_check_db_name") {
		if d.HasChange("global_kms_decrypt") {
			fullKmsDecryption := d.Get("global_kms_decrypt").(bool)
			if fullKmsDecryption {
				err := client.UpdateDatastoreGlobalKMSEncrypt(datastoreId, api.DataStoreInfra{FullKMSDecryption: fullKmsDecryption})
				if err != nil {
					return diag.FromErr(err)
				}
			} else {
				return diag.Errorf("At the moment you cannot deactivate global_kms_decrypt once it is set to true. You can message the Formal team for assistance.")
			}
		}
		if d.HasChange("username") || d.HasChange("password") {
			username := d.Get("username").(string)
			password := d.Get("password").(string)
			err := client.UpdateDatastoreUsernamePassword(datastoreId, api.DataStoreInfra{Username: username, Password: password})
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if d.HasChange("name") {
			name := d.Get("name").(string)
			err := client.UpdateDatastoreName(datastoreId, api.DataStoreInfra{Name: name})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("health_check_db_name") {
			healthCheckName := d.Get("health_check_db_name").(string)
			err := client.UpdateDatastoreHealthCheckDbName(datastoreId, api.DataStoreInfra{HealthCheckDbName: healthCheckName})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		return diag.Errorf("At the moment you can only update a datastore's global_kms_decrypt, username, password, and name. Please message the Formal team and we're happy to help.")

	}

	resourceDatastoreRead(ctx, d, meta)

	return diags
}

func resourceDatastoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dsId := d.Id()

	err := client.DeleteDatastore(dsId)
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err := client.GetDatastore(dsId)
		if err != nil {
			if strings.Contains(fmt.Sprint(err), "status: 404") {
				// Datastore was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on DatastoreStatus: ", map[string]interface{}{"err": err})
				currentErrors += 1
			}
		}

		if time.Since(deleteTimeStart) > time.Minute*15 {
			newErr := errors.New("deletion of this sidecar has taken more than 10m; the sidecar may be unhealthy")
			return diag.FromErr(newErr)
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")
	return diags
}
