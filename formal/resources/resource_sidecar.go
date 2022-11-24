package resource

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSidecar() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Sidecar with Formal.",
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
				Description: "Friendly name for this Sidecar.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Hostname of the original datastore.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the original datastore: supported values are `snowflake`, `postgres`, and `redshift`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"deployment_type": {
				// This description is used by the documentation generator and the language server.
				Description: "How the Sidecar should be deployed: `saas`, `managed`, or `onprem`.",
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
			"default_access_behavior": {
				// This description is used by the documentation generator and the language server.
				Description: "The default access behavior of the sidecar. Possible values are `allow` and `block`",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}
