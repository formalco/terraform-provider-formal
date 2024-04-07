package resource

import (
	"errors"
	"strconv"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"

	"github.com/formalco/terraform-provider-formal/formal/clients"

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
			"technology": {
				// This description is used by the documentation generator and the language server.
				Description: "Technology of the Datastore: supported values are`snowflake`, `postgres`, `redshift`, `mysql`, `mariadb`, `s3`, `dynamodb`, `mongodb`, `documentdb`, `http` and `ssh`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Creation time of the sidecar.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The hostname of the created sidecar.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
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

	sidecarReq := &corev1.CreateSidecarRequest{
		Name:                  d.Get("name").(string),
		Hostname:              d.Get("hostname").(string),
		Technology:            d.Get("technology").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.SidecarServiceClient.CreateSidecar(ctx, connect.NewRequest(sidecarReq))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdSidecar, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecar(ctx, connect.NewRequest(&corev1.GetSidecarRequest{Id: res.Msg.Sidecar.Id}))
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

	d.SetId(res.Msg.Sidecar.Id)

	resourceSidecarRead(ctx, d, meta)

	return diags
}

func resourceSidecarRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	sidecarId := d.Id()

	res, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecar(ctx, connect.NewRequest(&corev1.GetSidecarRequest{Id: sidecarId}))
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
	d.Set("formal_hostname", res.Msg.Sidecar.Hostname)
	d.Set("created_at", res.Msg.Sidecar.CreatedAt.AsTime().Unix())
	d.Set("technology", res.Msg.Sidecar.Technology)
	d.Set("termination_protection", res.Msg.Sidecar.TerminationProtection)

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

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecar(ctx, connect.NewRequest(&corev1.UpdateSidecarRequest{Id: sidecarId, Name: &name}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("name").(bool)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecar(ctx, connect.NewRequest(&corev1.UpdateSidecarRequest{Id: sidecarId, TerminationProtection: &terminationProtection}))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("hostname") {
		hostname := d.Get("name").(string)
		_, err := c.Grpc.Sdk.SidecarServiceClient.UpdateSidecar(ctx, connect.NewRequest(&corev1.UpdateSidecarRequest{Id: sidecarId, Hostname: &hostname}))
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

	_, err := c.Grpc.Sdk.SidecarServiceClient.DeleteSidecar(ctx, connect.NewRequest(&corev1.DeleteSidecarRequest{Id: dsId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err = c.Grpc.Sdk.SidecarServiceClient.GetSidecar(ctx, connect.NewRequest(&corev1.GetSidecarRequest{Id: dsId}))
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
		res, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecar(ctx, connect.NewRequest(&corev1.GetSidecarRequest{Id: val.(string)}))
		if err != nil {
			return nil, err
		}
		rawState["technology"] = res.Msg.Sidecar.Technology
	}

	return rawState, nil
}
