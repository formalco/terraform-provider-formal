package resource

import (
	"errors"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strconv"
	"time"

	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ResourceSidecar() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Sidecar with Formal.",
		CreateContext: resourceSidecarCreate,
		ReadContext:   resourceSidecarRead,
		UpdateContext: resourceSidecarUpdate,
		DeleteContext: resourceSidecarDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this Sidecar.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Datastore ID that the new Sidecar will be attached to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this Sidecar.",
				Type:        schema.TypeString,
				Required:    true,
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
				Description: "Configure DNS failover from the sidecar to the original datastore. In the unlikely case where the sidecar is unhealthy, having this value of `true` will forward traffic to the original database. Default `false`.",
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
			"cloud_provider": {
				// This description is used by the documentation generator and the language server.
				Description: "Cloud Provider that the sidecar sholud deploy in. Supported values at the moment are `aws`.",
				Type:        schema.TypeString,
				Required:    true,
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
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the Datastore: supported values are `snowflake`, `postgres`, `redshift` and `s3`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"formal_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The hostname of the created sidcar.",
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
				Required:    true,
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
			"version": {
				// This description is used by the documentation generator and the language server.
				Description: "Version of the Sidecar.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceSidecarCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	newSidecar := api.SidecarV2{
		Name:              d.Get("name").(string),
		DsId:              d.Get("datastore_id").(string),
		DataplaneId:       d.Get("dataplane_id").(string),
		CloudProvider:     d.Get("cloud_provider").(string),
		CloudRegion:       d.Get("cloud_region").(string),
		DeploymentType:    d.Get("deployment_type").(string),
		CloudAccountId:    d.Get("cloud_account_id").(string),
		FailOpen:          d.Get("fail_open").(bool),
		NetworkType:       d.Get("network_type").(string),
		FullKMSDecryption: d.Get("global_kms_decrypt").(bool),
		Version:           d.Get("version").(string),
		Technology:        d.Get("technology").(string),
	}

	sidecarId, err := c.Http.CreateSidecar(newSidecar)
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdSidecar, err := c.Http.GetSidecar(sidecarId)
		if err != nil {
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors+1)+" retrieving Sidecar: ", map[string]interface{}{"err": err})
				currentErrors += 1
				time.Sleep(15 * time.Second)
				continue
			}
		}

		if createdSidecar == nil {
			err = errors.New("sidecar with the given ID not found. It may have been deleted")
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "Sidecar Deployed state is: "+fmt.Sprint(createdSidecar.Deployed))
		// Check status
		if createdSidecar.Deployed {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	d.SetId(sidecarId)

	resourceSidecarRead(ctx, d, meta)

	return diags
}

func resourceSidecarRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarId := d.Id()

	sidecar, err := c.Http.GetSidecar(sidecarId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			tflog.Warn(ctx, "The Sidecar was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", sidecar.Id)
	d.Set("datastore_id", sidecar.DsId)
	d.Set("name", sidecar.Name)
	d.Set("formal_hostname", sidecar.FormalHostname)
	d.Set("cloud_provider", sidecar.CloudProvider)
	d.Set("cloud_region", sidecar.CloudRegion)
	d.Set("deployment_type", sidecar.DeploymentType)
	d.Set("cloud_account_id", sidecar.CloudAccountId)
	d.Set("fail_open", sidecar.FailOpen)
	d.Set("network_type", sidecar.NetworkType)
	d.Set("created_at", sidecar.CreatedAt)
	d.Set("global_kms_decrypt", sidecar.FullKMSDecryption)
	d.Set("dataplane_id", sidecar.DataplaneId)
	d.Set("version", sidecar.Version)

	if sidecar.DeploymentType == "onprem" {
		tlsCert, err := c.Http.GetSidecarTlsCert(sidecarId)
		if err != nil {
			return diag.FromErr(err)
		}
		if *tlsCert == "" {
			return diag.Errorf("The TLS Certificate was not found. Please contact the Formal team for support.")
		}

		d.Set("formal_control_plane_tls_certificate", *tlsCert)
	}

	d.SetId(sidecar.Id)

	return diags
}

func resourceSidecarUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarId := d.Id()

	fieldsThatCanChange := []string{"global_kms_decrypt", "name", "version"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	// Only enable updates to these fields, err otherwise
	if d.HasChange("global_kms_decrypt") {
		fullKmsDecryption := d.Get("global_kms_decrypt").(bool)
		if fullKmsDecryption {
			err := c.Http.UpdateSidecarGlobalKMSEncrypt(sidecarId, api.SidecarV2{FullKMSDecryption: fullKmsDecryption})
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.Errorf("At the moment you cannot deactivate global_kms_decrypt once it is set to true. You can message the Formal team for assistance.")
		}
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		err := c.Http.UpdateSidecarName(sidecarId, api.SidecarV2{Name: name})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("version") {
		version := d.Get("version").(string)
		err := c.Http.UpdateSidecarVersion(sidecarId, version)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceSidecarRead(ctx, d, meta)

	return diags
}

func resourceSidecarDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dsId := d.Id()

	err := c.Http.DeleteSidecar(dsId)
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err := c.Http.GetSidecar(dsId)
		if err != nil {
			if strings.Contains(fmt.Sprint(err), "status: 404") {
				// Sidecar was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on Sidecar Status: ", map[string]interface{}{"err": err})
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
