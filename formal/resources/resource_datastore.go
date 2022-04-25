package resource

import (
	"context"
	"time"

	"github.com/formalco/terraform-provider-formal/formal/api"
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
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the datastore: supported values are `snowflake`, `postgres`, and `redshift`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"deployment_type": {
				// This description is used by the documentation generator and the language server.
				Description: "How the sidecar for this datastore should be deployed: `saas`, `managed`, or `onprem`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fail_open": {
				// This description is used by the documentation generator and the language server.
				Description: "Configure DNS failover from the sidecar to the original datastore. In the unlikely case where the sidecar is healthy, having this value of `true` will forward traffic to the original database. Default `false`.",
				Type:        schema.TypeBool,
				Required:    true,
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
			"cloud_provider": {
				// This description is used by the documentation generator and the language server.
				Description: "Cloud Provider that the sidecar sholud deploy in. Supported values at the moment are `aws`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port your datastore is listening on. Required if your `technology` is `postgres` or `redshift`.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "The cloud region the sidecar should be deployed in. Supported values are `eu-west-1`, `eu-west-2`, `eu-west-3`,`eu-central-1`, `us-east-1`, `us-east-2`, `us-west-1`, `us-west-2`, `ap-southeast-1`",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cloud_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Required for managed cloud - the Formal ID for the connected Cloud Account. You can find this after creating the connection in the Formal Console.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"customer_vpc_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Required for managed cloud -- the VPC ID of the datastore.",
				Type:        schema.TypeString,
				Optional:    true,
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
		Name:     d.Get("name").(string),
		Hostname: d.Get("hostname").(string),
		Port:     portInt,
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
		// FormalHostName
		Technology:     d.Get("technology").(string),
		CloudProvider:  d.Get("cloud_provider").(string),
		CloudRegion:    d.Get("cloud_region").(string),
		DeploymentType: d.Get("deployment_type").(string),
		CloudAccountID: d.Get("cloud_account_id").(string),
		CustomerVpcId:  d.Get("customer_vpc_id").(string),
		// NetStackId:
		FailOpen: d.Get("fail_open").(bool),
		// CreateAt
	}

	res, err := client.CreateDatastore(newDatastore)
	if err != nil {
		return diag.FromErr(err)
	}

	for {
		createdDatastore, err := client.GetDatastoreStatus(res.DsId)
		if err != nil {
			return diag.FromErr(err)
		}
		if createdDatastore.ProxyStatus == "healthy" {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
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
		return diag.FromErr(err)
	}

	d.Set("id", datastore.Id)
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
	d.Set("customer_vpc_id", datastore.CustomerVpcId)
	d.Set("net_stack_id", datastore.NetStackId)
	d.Set("fail_open", datastore.FailOpen)
	d.Set("created_at", datastore.CreatedAt)

	// DsId is the UUID type id. See GetDatastoreInfraByDatastoreID in admin-api for more details
	d.SetId(datastore.DsId)

	return diags
}

func resourceDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Datastores are immutable. Please register a new datastore.")
	// TODO: implement fully, handle changes

	// client := meta.(*Client)
	// var diags diag.Diagnostics

	// datastoreId := d.Id()

	// if d.HasChange("hostname") {
	// d.Set("last_updated", time.Now().Format(time.RFC850))
	// /sidecar-hostname
	// }

	// if d.HasChange("username") || d.HasChange("password") {
	// 	d.Set("last_updated", time.Now().Format(time.RFC850))
	// 	// /credentials
	// }

	// return resourceDatastoreRead(ctx, d, meta)
}

func resourceDatastoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orderID := d.Id()

	err := client.DeleteDatastore(orderID)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: Parse for mark as deleted? check render

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}
