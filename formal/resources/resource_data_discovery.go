package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/robfig/cron/v3"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceDataDiscovery() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Data Discovery with Formal.",
		CreateContext: resourceDataDiscoveryCreate,
		ReadContext:   resourceDataDiscoveryRead,
		UpdateContext: resourceDataDiscoveryUpdate,
		DeleteContext: resourceDataDiscoveryDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID linked to this Data Discovery.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"native_user_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Native user ID linked to this Data Discovery.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Creation time of the Data Discovery.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"schedule": {
				// This description is used by the documentation generator and the language server.
				Description: "Schedule at which the Data Discovery will be executed. Possible values: `6h`, `12h`, `18h`, `24h` or a valid cron expression, for example `0 4,16 * * *` to run daily at 04:00 and 16:00 UTC.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					predefinedSchedules := map[string]bool{
						"6h":  true,
						"12h": true,
						"18h": true,
						"24h": true,
					}
					if predefinedSchedules[v] {
						return
					}
					if _, err := cron.ParseStandard(v); err != nil {
						errs = append(errs, fmt.Errorf("%q must be a valid cron expression or one of the predefined schedules ('6h', '12h', '18h', '24h')", key))
					}
					return
				},
			},
			"deletion_policy": {
				// This description is used by the documentation generator and the language server.
				Description: "Deletion policy of the Data Discovery. Possible values: `delete`, `mark_for_deletion`.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"delete",
					"mark_for_deletion",
				}, false),
			},
			"path": {
				// This description is used by the documentation generator and the language server.
				Description: "Path of the inventory object.",
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
			},
		},
	}
}

func resourceDataDiscoveryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	ResourceId := d.Get("resource_id").(string)
	NativeUserId := d.Get("native_user_id").(string)
	Schedule := d.Get("schedule").(string)
	DeletionPolicy := d.Get("deletion_policy").(string)
	Path := d.Get("path").(string)

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.CreateDataDiscoveryConfigurationRequest{
		ResourceId:     ResourceId,
		NativeUserId:   NativeUserId,
		Schedule:       Schedule,
		DeletionPolicy: DeletionPolicy,
		Path:           &Path,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.DataDiscoveryConfiguration.Id)

	resourceDataDiscoveryRead(ctx, d, meta)

	return diags
}

func resourceDataDiscoveryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := corev1.GetDataDiscoveryConfigurationRequest_DataDiscoveryConfigurationId{
		DataDiscoveryConfigurationId: d.Id(),
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.GetDataDiscoveryConfigurationRequest{Id: &id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.DataDiscoveryConfiguration.Id)

	return diags
}

func resourceDataDiscoveryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	dataDiscoveryId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"native_user_id", "schedule", "deletion_policy", "path"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	nativeUserId := d.Get("native_user_id").(string)
	schedule := d.Get("schedule").(string)
	deletionPolicy := d.Get("deletion_policy").(string)
	path := d.Get("path").(string)

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.UpdateDataDiscoveryConfigurationRequest{
		Id:             dataDiscoveryId,
		NativeUserId:   nativeUserId,
		Schedule:       schedule,
		DeletionPolicy: deletionPolicy,
		Path:           &path,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceDataDiscoveryRead(ctx, d, meta)

	return diags
}

func resourceDataDiscoveryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dataDiscoveryId := d.Id()

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.DeleteDataDiscoveryConfigurationRequest{Id: dataDiscoveryId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
