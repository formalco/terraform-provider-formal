package resource

import (
	"context"
	"fmt"
	"net/mail"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePolicyInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePolicyStateUpgradeV0,
			},
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
			"status": {
				// This description is used by the documentation generator and the language server.
				Description: "Defines the current status of the policy. It can be one of the following: 'draft', 'dry-run', or 'active'.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"draft",
					"dry-run",
					"active",
				}, false),
			},
			"notification": {
				// This description is used by the documentation generator and the language server.
				Description: "Notification settings for this policy.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"owner": {
				// This description is used by the documentation generator and the language server.
				Description: "Owner of this policy: it can be either a group name or a user email.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Policy cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			c := meta.(*clients.Clients)

			tflog.Debug(ctx, "Validating policy code", map[string]interface{}{
				"id":                 d.Id(),
				"has_module_changes": d.HasChange("module"),
			})

			if d.Id() == "" || d.HasChange("module") {
				resp, err := c.Grpc.Sdk.PoliciesServiceClient.GetPolicyCodeValidity(ctx, connect.NewRequest(&corev1.GetPolicyCodeValidityRequest{
					Code: d.Get("module").(string),
				}))
				if err != nil {
					return fmt.Errorf("policy code validation failed: %v", err)
				}
				if !resp.Msg.Valid {
					return fmt.Errorf("invalid policy code: %s", resp.Msg.Error)
				}
				tflog.Debug(ctx, "Policy code validation successful")
			}
			return nil
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	Name := d.Get("name").(string)
	Description := d.Get("description").(string)
	Module := d.Get("module").(string)
	Notification := d.Get("notification").(string)
	Status := d.Get("status").(string)
	TerminationProtection := d.Get("termination_protection").(bool)

	newPolicy := &corev1.CreatePolicyRequest{
		Name:                  Name,
		Description:           Description,
		Code:                  Module,
		Notification:          Notification,
		Status:                Status,
		TerminationProtection: TerminationProtection,
	}

	owner := d.Get("owner").(string)
	parsedEmail, _ := mail.ParseAddress(owner)
	if parsedEmail != nil {
		newPolicy.Owner = &corev1.CreatePolicyRequest_UserEmail{
			UserEmail: owner,
		}
	} else {
		newPolicy.Owner = &corev1.CreatePolicyRequest_GroupName{
			GroupName: owner,
		}
	}

	res, err := c.Grpc.Sdk.PoliciesServiceClient.CreatePolicy(ctx, connect.NewRequest(newPolicy))
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

	res, err := c.Grpc.Sdk.PoliciesServiceClient.GetPolicy(ctx, connect.NewRequest(&corev1.GetPolicyRequest{Id: policyId}))
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
	d.Set("status", res.Msg.Policy.Status)
	d.Set("termination_protection", res.Msg.Policy.TerminationProtection)

	d.SetId(policyId)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	policyId := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("module") || d.HasChange("notification") || d.HasChange("owner") || d.HasChange("active") || d.HasChange("status") || d.HasChange("termination_protection") {
		Name := d.Get("name").(string)
		Description := d.Get("description").(string)
		Module := d.Get("module").(string)
		Notification := d.Get("notification").(string)
		Status := d.Get("status").(string)
		TerminationProtection := d.Get("termination_protection").(bool)

		updatedPolicy := &corev1.UpdatePolicyRequest{
			Id:                    policyId,
			Name:                  Name,
			Description:           Description,
			Code:                  Module,
			Notification:          Notification,
			Status:                Status,
			TerminationProtection: TerminationProtection,
		}

		owner := d.Get("owner").(string)
		parsedEmail, _ := mail.ParseAddress(owner)
		if parsedEmail != nil {
			updatedPolicy.Owner = &corev1.UpdatePolicyRequest_UserEmail{
				UserEmail: owner,
			}
		} else {
			updatedPolicy.Owner = &corev1.UpdatePolicyRequest_GroupName{
				GroupName: owner,
			}
		}

		_, err := c.Grpc.Sdk.PoliciesServiceClient.UpdatePolicy(ctx, connect.NewRequest(updatedPolicy))

		if err != nil {
			return diag.FromErr(err)
		}

		return resourcePolicyRead(ctx, d, meta)
	} else {
		return diag.Errorf("At the moment you can only update a policy's name, description, module, notification, owner and active status. Please delete and recreate the Policy")
	}
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	policyId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Policy cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.PoliciesServiceClient.DeletePolicy(ctx, connect.NewRequest(&corev1.DeletePolicyRequest{Id: policyId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourcePolicyInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"status": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourcePolicyStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, fmt.Errorf("sidecar resource state upgrade failed, state is nil")
	}

	c := meta.(*clients.Clients)

	if val, ok := rawState["id"]; ok {
		res, err := c.Grpc.Sdk.PoliciesServiceClient.GetPolicy(ctx, connect.NewRequest(&corev1.GetPolicyRequest{Id: val.(string)}))
		if err != nil {
			return nil, err
		}
		rawState["status"] = res.Msg.Policy.Status
	}

	return rawState, nil
}
