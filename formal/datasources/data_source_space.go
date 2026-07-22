package datasources

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func Space() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for looking up a Space by ID or by name. Use either `id` or `name`, but not both.",
		ReadContext: spaceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The Formal ID for this Space.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Description:  "The name of the Space to look up. Use this to fetch a space by name.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"description": {
				Description: "Description of the Space.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Creation time of the Space.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"termination_protection": {
				Description: "If set to true, this Space cannot be deleted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func spaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var space *corev1.Space

	if spaceID, ok := d.GetOk("id"); ok {
		res, err := c.Grpc.Sdk.SpaceServiceClient.GetSpace(ctx, &corev1.GetSpaceRequest{Id: spaceID.(string)})
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				return diag.Errorf("no space found with id %s", spaceID)
			}
			return diag.FromErr(err)
		}
		space = res.Space
	} else {
		nameStr := d.Get("name").(string)
		res, err := c.Grpc.Sdk.SpaceServiceClient.ListSpaces(ctx, &corev1.ListSpacesRequest{
			Search:       nameStr,
			SearchFields: []string{"name"},
			Limit:        100,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		var foundSpace *corev1.Space
		for _, s := range res.Spaces {
			if s.Name == nameStr {
				foundSpace = s
				break
			}
		}
		if foundSpace == nil {
			return diag.Errorf("no space found with name %s", nameStr)
		}
		space = foundSpace
	}

	d.SetId(space.Id)
	d.Set("name", space.Name)
	d.Set("description", space.Description)
	d.Set("termination_protection", space.TerminationProtection)
	if space.CreatedAt != nil {
		d.Set("created_at", space.CreatedAt.AsTime().Unix())
	}

	return diags
}
