package resource

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const workflowSecretListPageSize = 500

func ResourceWorkflowSecret() *schema.Resource {
	return &schema.Resource{
		Description: "A secret that a workflow references as `secrets.<name>`. Only the workflow it belongs to can read it. The value is encrypted at rest and never returned by the API.",

		CreateContext: resourceWorkflowSecretCreate,
		ReadContext:   resourceWorkflowSecretRead,
		UpdateContext: resourceWorkflowSecretUpdate,
		DeleteContext: resourceWorkflowSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWorkflowSecretImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the workflow secret.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"workflow_id": {
				Description: "The ID of the workflow that can reference the secret.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:  "The name workflows reference the secret by, as `secrets.<name>`. Must be unique within the workflow, start with a letter or underscore, and contain only letters, digits and underscores.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,127}$`), "must start with a letter or underscore, contain only letters, digits and underscores, and be at most 128 characters"),
			},
			"value": {
				Description:  "The value of the secret. The API never returns it, so it is not read back. Prefer `value_wo` to keep the secret out of Terraform state.",
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				ExactlyOneOf: []string{"value", "value_wo"},
			},
			"value_wo": {
				Description:  "Write-only value of the secret. This value is not stored in Terraform state, so it must stay in the configuration. Requires Terraform 1.11+ and `value_wo_version`.",
				Type:         schema.TypeString,
				Optional:     true,
				WriteOnly:    true,
				ExactlyOneOf: []string{"value", "value_wo"},
				RequiredWith: []string{"value_wo_version"},
			},
			"value_wo_version": {
				Description:  "Version trigger for `value_wo`. Increment this value to update the secret.",
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"value_wo"},
			},
			"created_at": {
				Description: "When the workflow secret was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the workflow secret was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// writeOnlyWorkflowSecretValue reads value_wo from the raw configuration,
// since write-only values are absent from state.
func writeOnlyWorkflowSecretValue(d *schema.ResourceData) (string, bool, error) {
	if d.GetRawConfig().IsNull() {
		return "", false, nil
	}
	value, diags := d.GetRawConfigAt(cty.GetAttrPath("value_wo"))
	if diags.HasError() {
		return "", false, fmt.Errorf("failed to get value_wo: %v", diags)
	}
	if value.IsNull() {
		return "", false, nil
	}
	if !value.IsKnown() || value.Type() != cty.String {
		return "", false, errors.New("value_wo must be a known string")
	}
	return value.AsString(), true, nil
}

func resourceWorkflowSecretCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	value := d.Get("value").(string)
	writeOnlyValue, configured, err := writeOnlyWorkflowSecretValue(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if configured {
		value = writeOnlyValue
	}

	res, err := c.Grpc.Sdk.WorkflowServiceClient.CreateWorkflowSecret(ctx, &corev1.CreateWorkflowSecretRequest{
		WorkflowId: d.Get("workflow_id").(string),
		Name:       d.Get("name").(string),
		Value:      value,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.WorkflowSecret.Id)
	return flattenWorkflowSecret(d, res.WorkflowSecret)
}

func resourceWorkflowSecretRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	secret, found, err := findWorkflowSecret(ctx, c, d.Get("workflow_id").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		tflog.Warn(ctx, "The workflow secret with ID "+d.Id()+" was not found, which means it may have been deleted without using this Terraform config.")
		d.SetId("")
		return nil
	}

	return flattenWorkflowSecret(d, secret)
}

func resourceWorkflowSecretUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var value string
	switch {
	case d.HasChange("value_wo_version"):
		writeOnlyValue, configured, err := writeOnlyWorkflowSecretValue(d)
		if err != nil {
			return diag.FromErr(err)
		}
		if !configured {
			return diag.Errorf("value_wo_version requires value_wo")
		}
		value = writeOnlyValue
	case d.HasChange("value"):
		value = d.Get("value").(string)
	default:
		return nil
	}

	res, err := c.Grpc.Sdk.WorkflowServiceClient.UpdateWorkflowSecret(ctx, &corev1.UpdateWorkflowSecretRequest{
		Id:    d.Id(),
		Value: value,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return flattenWorkflowSecret(d, res.WorkflowSecret)
}

func resourceWorkflowSecretDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.WorkflowServiceClient.DeleteWorkflowSecret(ctx, &corev1.DeleteWorkflowSecretRequest{Id: d.Id()})
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// The API lists secrets per workflow and has no get by ID, so secrets are
// imported as <workflow_id>/<secret_id>.
func resourceWorkflowSecretImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	workflowID, secretID, ok := strings.Cut(d.Id(), "/")
	if !ok || workflowID == "" || secretID == "" {
		return nil, errors.New("import ID must be <workflow_id>/<secret_id>")
	}
	d.SetId(secretID)
	d.Set("workflow_id", workflowID)

	if diags := resourceWorkflowSecretRead(ctx, d, meta); diags.HasError() {
		return nil, errors.New(diags[0].Summary)
	}
	if d.Id() == "" {
		return nil, errors.New("workflow secret " + secretID + " not found in workflow " + workflowID)
	}
	return []*schema.ResourceData{d}, nil
}

func findWorkflowSecret(ctx context.Context, c *clients.Clients, workflowID, id string) (*corev1.WorkflowSecret, bool, error) {
	cursor := ""
	for {
		res, err := c.Grpc.Sdk.WorkflowServiceClient.ListWorkflowSecrets(ctx, &corev1.ListWorkflowSecretsRequest{
			WorkflowId: workflowID,
			Limit:      workflowSecretListPageSize,
			Cursor:     cursor,
		})
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				return nil, false, nil
			}
			return nil, false, err
		}
		for _, secret := range res.WorkflowSecrets {
			if secret.Id == id {
				return secret, true, nil
			}
		}
		if res.ListMetadata.GetNextCursor() == "" {
			return nil, false, nil
		}
		cursor = res.ListMetadata.GetNextCursor()
	}
}

func flattenWorkflowSecret(d *schema.ResourceData, secret *corev1.WorkflowSecret) diag.Diagnostics {
	d.Set("workflow_id", secret.WorkflowId)
	d.Set("name", secret.Name)
	d.Set("created_at", secret.CreatedAt.AsTime().UTC().Format(time.RFC3339))
	d.Set("updated_at", secret.UpdatedAt.AsTime().UTC().Format(time.RFC3339))
	return nil
}
