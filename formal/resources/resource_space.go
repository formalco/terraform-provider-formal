package resource

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSpace() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Space with Formal.",
		CreateContext: resourceSpaceCreate,
		ReadContext:   resourceSpaceRead,
		UpdateContext: resourceSpaceUpdate,
		DeleteContext: resourceSpaceDelete,
		SchemaVersion: 1,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this Space.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this Space.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Description of the Space.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Creation time of the Space.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Space cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	spaceReq := &corev1.CreateSpaceRequest{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.SpaceServiceClient.CreateSpace(ctx, spaceReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Space.Id)

	resourceSpaceRead(ctx, d, meta)

	return diags
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	spaceId := d.Id()

	res, err := c.Grpc.Sdk.SpaceServiceClient.GetSpace(ctx, &corev1.GetSpaceRequest{Id: spaceId})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Space was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Space.Id)
	d.Set("name", res.Space.Name)
	d.Set("description", res.Space.Description)
	d.Set("created_at", res.Space.CreatedAt.AsTime().Unix())
	d.Set("termination_protection", res.Space.TerminationProtection)

	d.SetId(res.Space.Id)

	return diags
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	spaceId := d.Id()

	fieldsThatCanChange := []string{"name", "description", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	name := d.Get("name").(string)
	terminationProtection := d.Get("termination_protection").(bool)
	description := d.Get("description").(string)

	_, err := c.Grpc.Sdk.SpaceServiceClient.UpdateSpace(ctx, &corev1.UpdateSpaceRequest{Id: spaceId, Name: &name, TerminationProtection: &terminationProtection, Description: &description})
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSpaceRead(ctx, d, meta)

	return diags
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	spaceId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Space cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.SpaceServiceClient.DeleteSpace(ctx, &corev1.DeleteSpaceRequest{Id: spaceId})
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	deleteTimeStart := time.Now().UTC()
	for {
		// Retrieve status
		_, err = c.Grpc.Sdk.SpaceServiceClient.GetSpace(ctx, &corev1.GetSpaceRequest{Id: spaceId})
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				tflog.Info(ctx, "Space deleted", map[string]any{"space_id": spaceId})
				// Space was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ErrorTolerance {
				return diag.FromErr(err)
			}
			tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on Space Status: ", map[string]any{"err": err})
			currentErrors++
		}

		if time.Since(deleteTimeStart) > time.Minute*15 {
			newErr := errors.New("deletion of this space has taken more than 15m")
			return diag.FromErr(newErr)
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")
	return diags
}
