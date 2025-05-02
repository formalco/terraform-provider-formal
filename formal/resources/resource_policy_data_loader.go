package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourcePolicyDataLoader() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a policy data loader with Formal.",
		CreateContext: resourcePolicyDataLoaderCreate,
		ReadContext:   resourcePolicyDataLoaderRead,
		UpdateContext: resourcePolicyDataLoaderUpdate,
		DeleteContext: resourcePolicyDataLoaderDelete,
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
				Description: "Id of this policy data loader.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this policy data loader.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Policy data loader description.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				// This description is used by the documentation generator and the language server.
				Description: "The key to access the output data of this policy data loader.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"worker_runtime": {
				// This description is used by the documentation generator and the language server.
				Description: "The execution environment for the code. It can be one of the following: 'python3.11' or 'nodejs18.x'.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"python3.11",
					"nodejs18.x",
				}, false),
			},
			"worker_code": {
				// This description is used by the documentation generator and the language server.
				Description: "The code that will be executed to fetch and output the data.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"worker_schedule": {
				// This description is used by the documentation generator and the language server.
				Description: "Second-based 'cron' expression specifying when the data should be fetched. For example, use '*/10 * * * * *' to run the code every 10 seconds.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				// This description is used by the documentation generator and the language server.
				Description: "Defines the current status of the policy data loader. It can be one of the following: 'draft' or 'active'.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"draft",
					"dry-run",
					"active",
				}, false),
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this policy data loader cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the policy data loader was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourcePolicyDataLoaderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	req := &corev1.CreatePolicyDataLoaderRequest{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		Key:                   d.Get("key").(string),
		WorkerRuntime:         d.Get("worker_runtime").(string),
		WorkerCode:            d.Get("worker_code").(string),
		WorkerSchedule:        d.Get("worker_schedule").(string),
		Status:                d.Get("status").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.PolicyDataLoaderServiceClient.CreatePolicyDataLoader(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.PolicyDataLoader.Id)
	return resourcePolicyDataLoaderRead(ctx, d, meta)
}

func resourcePolicyDataLoaderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	loaderId := d.Id()

	res, err := c.Grpc.Sdk.PolicyDataLoaderServiceClient.GetPolicyDataLoader(ctx, connect.NewRequest(&corev1.GetPolicyDataLoaderRequest{Id: loaderId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Loader was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.PolicyDataLoader.Id)
	d.Set("name", res.Msg.PolicyDataLoader.Name)
	d.Set("description", res.Msg.PolicyDataLoader.Description)
	d.Set("key", res.Msg.PolicyDataLoader.Key)
	d.Set("worker_runtime", res.Msg.PolicyDataLoader.WorkerRuntime)
	d.Set("worker_code", res.Msg.PolicyDataLoader.WorkerCode)
	d.Set("worker_schedule", res.Msg.PolicyDataLoader.WorkerSchedule)
	d.Set("status", res.Msg.PolicyDataLoader.Status)
	d.Set("termination_protection", res.Msg.PolicyDataLoader.TerminationProtection)
	d.Set("created_at", res.Msg.PolicyDataLoader.CreatedAt)
	d.Set("updated_at", res.Msg.PolicyDataLoader.UpdatedAt)

	d.SetId(res.Msg.PolicyDataLoader.Id)
	return diags
}

func resourcePolicyDataLoaderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	loaderId := d.Id()

	fieldsThatCanChange := []string{"name", "description", "key", "worker_runtime", "worker_code", "worker_schedule", "status", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	key := d.Get("key").(string)
	workerRuntime := d.Get("worker_runtime").(string)
	workerCode := d.Get("worker_code").(string)
	workerSchedule := d.Get("worker_schedule").(string)
	status := d.Get("status").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.Sdk.PolicyDataLoaderServiceClient.UpdatePolicyDataLoader(ctx, connect.NewRequest(&corev1.UpdatePolicyDataLoaderRequest{
		Id:                    loaderId,
		Name:                  &name,
		Description:           &description,
		Key:                   &key,
		WorkerRuntime:         &workerRuntime,
		WorkerCode:            &workerCode,
		WorkerSchedule:        &workerSchedule,
		Status:                &status,
		TerminationProtection: &terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcePolicyDataLoaderRead(ctx, d, meta)
}

func resourcePolicyDataLoaderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	loaderId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Policy data loader cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.PolicyDataLoaderServiceClient.DeletePolicyDataLoader(ctx, connect.NewRequest(&corev1.DeletePolicyDataLoaderRequest{Id: loaderId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
