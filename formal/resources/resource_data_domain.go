package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceDataDomain() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Registering a Data Domain with Formal.",

		CreateContext: resourceDataDomainCreate,
		ReadContext:   resourceDataDomainRead,
		DeleteContext: resourceDataDomainDelete,
		UpdateContext: resourceDataDomainUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePolicyInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePolicyStateUpgradeV0,
			},
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Id of this data domain.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Name of the data domain.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Description of the data domain.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the policy was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"included_paths": {
				// This description is used by the documentation generator and the language server.
				Description: "Included paths of this data domain.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"excluded_paths": {
				// This description is used by the documentation generator and the language server.
				Description: "Excluded paths of this data domain.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"owners": {
				// This description is used by the documentation generator and the language server.
				Description: "Owners of this policy.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceDataDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	ownersInterface := d.Get("owners").([]interface{}) // Retrieve the interface slice.
	var Owners []*corev1.DomainOwner
	for _, ownerInterface := range ownersInterface {
		ownerMap, ok := ownerInterface.(map[string]interface{}) // Perform type assertion.
		if !ok {
			// Handle the error where the type assertion fails.
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error reading owner",
				Detail:   "An unexpected type was encountered while reading an owner, expected 'map[string]interface{}'",
			})
			continue
		}

		// Assuming 'object_type' and 'object_id' are the keys and they are of type string.
		// You need to adjust the logic below if the structure is different or more complex.
		owner := corev1.DomainOwner{
			ObjectType: ownerMap["object_type"].(string), // Direct type assertion; consider error handling here.
			ObjectId:   ownerMap["object_id"].(string),   // Direct type assertion; consider error handling here.
		}
		Owners = append(Owners, &owner)
	}

	// Maps to user-defined fields
	Name := d.Get("name").(string)
	Description := d.Get("description").(string)

	var IncludedPaths []string
	for _, includedPath := range d.Get("included_paths").([]interface{}) {
		IncludedPaths = append(IncludedPaths, includedPath.(string))
	}

	var ExcludedPaths []string
	for _, excludedPath := range d.Get("excluded_paths").([]interface{}) {
		ExcludedPaths = append(ExcludedPaths, excludedPath.(string))
	}

	res, err := c.Grpc.Sdk.InventoryServiceClient.CreateDataDomain(ctx, connect.NewRequest(&corev1.CreateDataDomainRequest{
		Name:          Name,
		Description:   Description,
		Owners:        Owners,
		IncludedPaths: IncludedPaths,
		ExcludedPaths: ExcludedPaths,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Domain.Id)

	resourceDataDomainRead(ctx, d, meta)
	return diags
}

func resourceDataDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	domainId := d.Id()

	res, err := c.Grpc.Sdk.InventoryServiceClient.GetDataDomain(ctx, connect.NewRequest(&corev1.GetDataDomainRequest{Id: domainId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Domain with ID "+domainId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of Policy
	d.Set("id", res.Msg.Domain.Id)
	d.Set("name", res.Msg.Domain.Name)
	d.Set("description", res.Msg.Domain.Description)
	d.Set("included_paths", res.Msg.Domain.IncludedPaths)
	d.Set("excluded_paths", res.Msg.Domain.ExcludedPaths)
	d.Set("owners", res.Msg.Domain.Owners)
	d.Set("created_at", res.Msg.Domain.CreatedAt.AsTime().Unix())

	d.SetId(domainId)

	return diags
}

func resourceDataDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	domainId := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("owners") || d.HasChange("included_paths") || d.HasChange("excluded_paths") {
		Name := d.Get("name").(string)
		Description := d.Get("description").(string)

		var IncludedPaths []string
		for _, includedPath := range d.Get("included_paths").([]interface{}) {
			IncludedPaths = append(IncludedPaths, includedPath.(string))
		}

		var ExcludedPaths []string
		for _, excludedPath := range d.Get("excluded_paths").([]interface{}) {
			ExcludedPaths = append(ExcludedPaths, excludedPath.(string))
		}

		ownersInterface := d.Get("owners").([]interface{}) // Retrieve the interface slice.
		var Owners []*corev1.DomainOwner
		for _, ownerInterface := range ownersInterface {
			ownerMap, ok := ownerInterface.(map[string]interface{}) // Perform type assertion.
			if !ok {
				continue
			}

			// Assuming 'object_type' and 'object_id' are the keys and they are of type string.
			// You need to adjust the logic below if the structure is different or more complex.
			owner := corev1.DomainOwner{
				ObjectType: ownerMap["object_type"].(string), // Direct type assertion; consider error handling here.
				ObjectId:   ownerMap["object_id"].(string),   // Direct type assertion; consider error handling here.
			}
			Owners = append(Owners, &owner)
		}
		_, err := c.Grpc.Sdk.InventoryServiceClient.UpdateDataDomain(ctx, connect.NewRequest(&corev1.UpdateDataDomainRequest{
			Id:            domainId,
			Name:          Name,
			Description:   Description,
			Owners:        Owners,
			IncludedPaths: IncludedPaths,
			ExcludedPaths: ExcludedPaths,
		}))

		if err != nil {
			return diag.FromErr(err)
		}

		return resourceDataDomainRead(ctx, d, meta)
	} else {
		return diag.Errorf("At the moment you can only update a policy's name, description, module, notification, owners and active status. Please delete and recreate the Policy")
	}
}

func resourceDataDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	domainId := d.Id()

	_, err := c.Grpc.Sdk.InventoryServiceClient.DeleteDataDomain(ctx, connect.NewRequest(&corev1.DeleteDataDomainRequest{Id: domainId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
