package resource

import (
	"errors"
	"strconv"
	"time"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/bufbuild/connect-go"

	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/formalco/terraform-provider-formal/formal/validation"

	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourceSidecarInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceSidecarStateUpgradeV0,
			},
		},

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
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this Sidecar.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Datastore ID that the new Sidecar will be attached to.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Deprecated:  "This field is deprecated. Please use formal_sidecar_datastore_link resource instead. This attribute will be removed in the next major version of the provider.",
			},
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description:  "Technology of the Datastore: supported values are`snowflake`, `postgres`, `redshift`, `mysql`, `mariadb`, `s3`, `http` and `ssh`.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.Technology(),
			},
			"deployment_type": {
				// This description is used by the documentation generator and the language server.
				Description:  "How the Sidecar should be deployed: `managed`, or `onprem`.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.SidecarDeploymentType(),
			},
			"fail_open": {
				// This description is used by the documentation generator and the language server.
				Description: "Configure DNS failover from the sidecar to the original datastore. In the unlikely case where the sidecar is unhealthy, having this value of `true` will forward traffic to the original database. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
			"network_type": {
				// This description is used by the documentation generator and the language server.
				Description:  "Configure the sidecar network type. Value can be `internet-facing`, `internal` or `internet-and-internal`.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.SidecarNetworkType(),
			},
			"formal_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The hostname of the created sidecar.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
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
				Description: "Version of the Sidecar to deploy for `managed`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Sidecar cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Api key for the deployed Sidecar.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceSidecarCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarReq := &adminv1.CreateSidecarRequest{
		Name:                  d.Get("name").(string),
		DataplaneId:           d.Get("dataplane_id").(string),
		DeploymentType:        d.Get("deployment_type").(string),
		FailOpen:              d.Get("fail_open").(bool),
		NetworkType:           d.Get("network_type").(string),
		GlobalKmsDecrypt:      d.Get("global_kms_decrypt").(bool),
		Version:               d.Get("version").(string),
		Technology:            d.Get("technology").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}
	hostname := d.Get("formal_hostname").(string)
	if sidecarReq.DeploymentType == "onprem" && hostname != "" {
		sidecarReq.FormalHostname = hostname
	}

	res, err := c.Grpc.Sdk.SidecarServiceClient.CreateSidecar(ctx, connect.NewRequest(sidecarReq))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdSidecar, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecarById(ctx, connect.NewRequest(&adminv1.GetSidecarByIdRequest{Id: res.Msg.Id}))
		if err != nil {
			if currentErrors >= ErrorTolerance {
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

		tflog.Info(ctx, "Sidecar Deployed state is: "+fmt.Sprint(createdSidecar.Msg.Sidecar.Deployed))
		// Check status
		if createdSidecar.Msg.Sidecar.Deployed {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	d.SetId(res.Msg.Id)

	resourceSidecarRead(ctx, d, meta)

	return diags
}

func resourceSidecarRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarId := d.Id()

	res, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecarById(ctx, connect.NewRequest(&adminv1.GetSidecarByIdRequest{Id: sidecarId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Sidecar was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Sidecar.Id)
	d.Set("name", res.Msg.Sidecar.Name)
	d.Set("formal_hostname", res.Msg.Sidecar.FormalHostname)
	d.Set("deployment_type", res.Msg.Sidecar.DeploymentType)
	d.Set("fail_open", res.Msg.Sidecar.FailOpen)
	d.Set("network_type", res.Msg.Sidecar.NetworkType)
	d.Set("created_at", res.Msg.Sidecar.CreatedAt.AsTime().Unix())
	d.Set("global_kms_decrypt", res.Msg.Sidecar.GlobalKmsDecrypt)
	d.Set("dataplane_id", res.Msg.Sidecar.DataplaneId)
	d.Set("version", res.Msg.Sidecar.Version)
	d.Set("technology", res.Msg.Sidecar.Technology)
	d.Set("termination_protection", res.Msg.Sidecar.TerminationProtection)

	if res.Msg.Sidecar.DeploymentType == "onprem" && c.Grpc.ReturnSensitiveValue {
		res, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecarTlsCertificateById(ctx, connect.NewRequest(&adminv1.GetSidecarTlsCertificateByIdRequest{Id: sidecarId}))
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("formal_control_plane_tls_certificate", res.Msg.Secret)
		d.Set("api_key", res.Msg.Secret)
	}

	d.SetId(res.Msg.Sidecar.Id)

	return diags
}

func resourceSidecarUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarId := d.Id()

	fieldsThatCanChange := []string{"global_kms_decrypt", "name", "version", "formal_hostname", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	// Only enable updates to these fields, err otherwise
	if d.HasChange("global_kms_decrypt") {
		fullKmsDecryption := d.Get("global_kms_decrypt").(bool)
		if fullKmsDecryption {
			_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecarKmsDecryptPolicy(ctx, connect.NewRequest(&adminv1.UpdateSidecarKmsDecryptPolicyRequest{Id: sidecarId, Enabled: true}))
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.Errorf("At the moment you cannot deactivate global_kms_decrypt once it is set to true. You can message the Formal team for assistance.")
		}
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecarName(ctx, connect.NewRequest(&adminv1.UpdateSidecarNameRequest{Id: sidecarId, Name: name}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("version") {
		version := d.Get("version").(string)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecarVersion(ctx, connect.NewRequest(&adminv1.UpdateSidecarVersionRequest{Id: sidecarId, Version: version}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("formal_hostname") {
		if d.Get("deployment_type").(string) != "onprem" {
			return diag.Errorf("formal_hostname can only be updated for onprem deployment types")
		}

		formalHostname := d.Get("formal_hostname").(string)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecarFormalHostname(ctx, connect.NewRequest(&adminv1.UpdateSidecarFormalHostnameRequest{Id: sidecarId, Hostname: formalHostname}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateTerminationProtection(ctx, connect.NewRequest(&adminv1.UpdateTerminationProtectionRequest{
			Id:      sidecarId,
			Enabled: terminationProtection,
		}))
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

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Sidecar cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.SidecarServiceClient.DeleteSidecar(ctx, connect.NewRequest(&adminv1.DeleteSidecarRequest{Id: dsId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err = c.Grpc.Sdk.SidecarServiceClient.GetSidecarById(ctx, connect.NewRequest(&adminv1.GetSidecarByIdRequest{Id: dsId}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				tflog.Info(ctx, "Sidecar deleted", map[string]interface{}{"sidecar_id": dsId})
				// Sidecar was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ErrorTolerance {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on Sidecar Status: ", map[string]interface{}{"err": err})
				currentErrors += 1
			}
		}

		if time.Since(deleteTimeStart) > time.Minute*15 {
			newErr := errors.New("deletion of this sidecar has taken more than 15m; the sidecar may be unhealthy")
			return diag.FromErr(newErr)
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")
	return diags
}

func resourceSidecarInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"technology": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSidecarStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, fmt.Errorf("sidecar resource state upgrade failed, state is nil")
	}

	c := meta.(*clients.Clients)

	if val, ok := rawState["id"]; ok {
		res, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecarById(ctx, connect.NewRequest(&adminv1.GetSidecarByIdRequest{Id: val.(string)}))
		if err != nil {
			return nil, err
		}
		rawState["technology"] = res.Msg.Sidecar.Technology
	}

	return rawState, nil
}
