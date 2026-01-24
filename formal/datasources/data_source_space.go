package datasources

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func Space() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for looking up a Space by name.",
		ReadContext: spaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Space.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The Formal ID for this Space.",
				Type:        schema.TypeString,
				Computed:    true,
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

func spaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	name, ok := d.GetOk("name")
	if !ok {
		return diag.Errorf("name must be specified")
	}
	nameStr := name.(string)

	// List spaces with search filter to narrow down results
	res, err := c.Grpc.Sdk.SpaceServiceClient.ListSpaces(ctx, connect.NewRequest(&corev1.ListSpacesRequest{
		Search:       nameStr,
		SearchFields: []string{"name"},
		Limit:        100,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	// Find exact match by name
	var foundSpace *corev1.Space
	for _, space := range res.Msg.Spaces {
		if space.Name == nameStr {
			foundSpace = space
			break
		}
	}

	if foundSpace == nil {
		return diag.Errorf("no space found with name %s", nameStr)
	}

	d.SetId(foundSpace.Id)
	d.Set("name", foundSpace.Name)
	d.Set("description", foundSpace.Description)
	d.Set("termination_protection", foundSpace.TerminationProtection)
	d.Set("created_at", foundSpace.CreatedAt.AsTime().Unix())

	return diags
}
