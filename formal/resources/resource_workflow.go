package resource

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceWorkflow() *schema.Resource {
	return &schema.Resource{
		Description: "Workflows enable automation of actions based on triggers. A workflow is defined using YAML code that specifies a trigger (what starts the workflow) and actions (what the workflow does).",

		CreateContext: resourceWorkflowCreate,
		ReadContext:   resourceWorkflowRead,
		UpdateContext: resourceWorkflowUpdate,
		DeleteContext: resourceWorkflowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the workflow.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the workflow. Must be unique within the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"code": {
				Description: "The workflow definition in YAML format. Defines the trigger and actions for the workflow.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "The workflow status. Accepted values are `active` and `draft`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "active",
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"draft",
				}, false),
			},
		},
	}
}

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	status := d.Get("status").(string)
	req := &corev1.CreateWorkflowRequest{
		Name:   d.Get("name").(string),
		Code:   d.Get("code").(string),
		Status: &status,
	}

	res, err := c.Grpc.Sdk.WorkflowServiceClient.CreateWorkflow(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Workflow.Id)

	return resourceWorkflowRead(ctx, d, meta)
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	workflowId := d.Id()

	res, err := c.Grpc.Sdk.WorkflowServiceClient.GetWorkflow(ctx, &corev1.GetWorkflowRequest{Id: workflowId})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Workflow with ID "+workflowId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Workflow.Id)
	d.Set("name", res.Workflow.Name)
	d.Set("code", res.Workflow.Code)
	d.Set("status", res.Workflow.GetStatus())

	return nil
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	if d.HasChange("name") || d.HasChange("code") || d.HasChange("status") {
		status := d.Get("status").(string)
		req := &corev1.UpdateWorkflowRequest{
			Id:     d.Id(),
			Name:   d.Get("name").(string),
			Code:   d.Get("code").(string),
			Status: &status,
		}

		_, err := c.Grpc.Sdk.WorkflowServiceClient.UpdateWorkflow(ctx, req)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceWorkflowRead(ctx, d, meta)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.WorkflowServiceClient.DeleteWorkflow(ctx, &corev1.DeleteWorkflowRequest{Id: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
