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
				Required:    true,
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

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	spaceReq := &corev1.CreateSpaceRequest{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.SpaceServiceClient.CreateSpace(ctx, connect.NewRequest(spaceReq))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Space.Id)

	resourceSpaceRead(ctx, d, meta)

	return diags
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	spaceId := d.Id()

	res, err := c.Grpc.Sdk.SpaceServiceClient.GetSpace(ctx, connect.NewRequest(&corev1.GetSpaceRequest{Id: spaceId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Space was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Space.Id)
	d.Set("name", res.Msg.Space.Name)
	d.Set("description", res.Msg.Space.Description)
	d.Set("created_at", res.Msg.Space.CreatedAt.AsTime().Unix())
	d.Set("termination_protection", res.Msg.Space.TerminationProtection)

	d.SetId(res.Msg.Space.Id)

	return diags
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	spaceId := d.Id()

	fieldsThatCanChange := []string{"name", "description", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	name := d.Get("name").(string)
	terminationProtection := d.Get("termination_protection").(bool)
	description := d.Get("description").(string)

	_, err := c.Grpc.Sdk.SpaceServiceClient.UpdateSpace(ctx, connect.NewRequest(&corev1.UpdateSpaceRequest{Id: spaceId, Name: &name, TerminationProtection: &terminationProtection, Description: &description}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSpaceRead(ctx, d, meta)

	return diags
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	spaceId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Space cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.SpaceServiceClient.DeleteSpace(ctx, connect.NewRequest(&corev1.DeleteSpaceRequest{Id: spaceId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err = c.Grpc.Sdk.SpaceServiceClient.GetSpace(ctx, connect.NewRequest(&corev1.GetSpaceRequest{Id: spaceId}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				tflog.Info(ctx, "Space deleted", map[string]interface{}{"space_id": spaceId})
				// Space was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ErrorTolerance {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on Space Status: ", map[string]interface{}{"err": err})
				currentErrors += 1
			}
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
