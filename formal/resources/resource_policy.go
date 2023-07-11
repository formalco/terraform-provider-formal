package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Policy in Formal.",

		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Policy Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Policy Description.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"module": {
				// This description is used by the documentation generator and the language server.
				Description: "The module describing how the policy works. Create one in the Formal Console.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of this Policy.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				// This description is used by the documentation generator and the language server.
				Description: "Who the policy was created by.",
				Type:        schema.TypeString,
				Computed:    true,
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
			"active": {
				// This description is used by the documentation generator and the language server.
				Description: "Active status of this policy.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"org_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for your organisation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expire_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When this policy is set to expire.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				// This description is used by the documentation generator and the language server.
				Description: "Additional descriptor for active status of this policy.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"notification": {
				// This description is used by the documentation generator and the language server.
				Description: "Notification settings for this policy.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"owners": {
				// This description is used by the documentation generator and the language server.
				Description: "Owners of this policy.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var owners []string
	for _, owner := range d.Get("owners").([]interface{}) {
		owners = append(owners, owner.(string))
	}

	// Maps to user-defined fields
	Name := d.Get("name").(string)
	Description := d.Get("description").(string)
	Module := d.Get("module").(string)
	SourceType := "terraform"
	Notification := d.Get("notification").(string)
	Active := d.Get("active").(bool)

	res, err := c.Grpc.Sdk.PolicyServiceClient.CreatePolicy(ctx, connect.NewRequest(&adminv1.CreatePolicyRequest{
		Name:         Name,
		Description:  Description,
		Code:         Module,
		Notification: Notification,
		Owners:       owners,
		SourceType:   SourceType,
		Active:       Active,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Policy.Id)

	resourcePolicyRead(ctx, d, meta)
	return diags
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	policyId := d.Id()

	res, err := c.Grpc.Sdk.PolicyServiceClient.GetPolicy(ctx, connect.NewRequest(&adminv1.GetPolicyRequest{PolicyId: policyId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Policy with ID "+policyId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of Policy
	d.Set("id", res.Msg.Policy.Id)
	d.Set("name", res.Msg.Policy.Name)
	d.Set("description", res.Msg.Policy.Description)
	d.Set("module", res.Msg.Policy.Code)
	d.Set("notification", res.Msg.Policy.Notification)
	d.Set("owners", res.Msg.Policy.Owners)
	d.Set("active", res.Msg.Policy.Active)

	d.SetId(policyId)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	policyId := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("module") || d.HasChange("notification") || d.HasChange("owners") || d.HasChange("active") {

		var owners []string
		for _, owner := range d.Get("owners").([]interface{}) {
			owners = append(owners, owner.(string))
		}

		Name := d.Get("name").(string)
		Description := d.Get("description").(string)
		Module := d.Get("module").(string)
		Notification := d.Get("notification").(string)
		SourceType := "terraform"
		Active := d.Get("active").(bool)

		_, err := c.Grpc.Sdk.PolicyServiceClient.UpdatePolicy(ctx, connect.NewRequest(&adminv1.UpdatePolicyRequest{
			Id:           policyId,
			SourceType:   SourceType,
			Name:         Name,
			Description:  Description,
			Code:         Module,
			Notification: Notification,
			Owners:       owners,
			Active:       Active,
		}))

		if err != nil {
			return diag.FromErr(err)
		}

		return resourcePolicyRead(ctx, d, meta)
	} else {
		return diag.Errorf("At the moment you can only update a policy's name, description, module, notification, owners and active status. Please delete and recreate the Policy")
	}
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	policyId := d.Id()

	_, err := c.Grpc.Sdk.PolicyServiceClient.DeletePolicy(ctx, connect.NewRequest(&adminv1.DeletePolicyRequest{Id: policyId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
