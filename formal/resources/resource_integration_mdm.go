package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceIntegrationMDM() *schema.Resource {
	return &schema.Resource{
		Description:        "Registering a Integration MDM app.",
		DeprecationMessage: "This resource is deprecated and will be removed in a future version. MDM integration support is being phased out.",
		CreateContext:      resourceIntegrationMDMCreate,
		ReadContext:        resourceIntegrationMDMRead,
		DeleteContext:      resourceIntegrationMDMDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the App.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Friendly name for the Integration app.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"kandji": {
				Description: "Configuration block for Kandji integration.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {
							Description: "API Key of your Kandji organization.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"api_url": {
							Description: "API URL of your Kandji organization.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceIntegrationMDMCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics
	var res *connect.Response[corev1.CreateIntegrationMDMResponse]
	var err error

	// Check if Kandji is configured
	if v, ok := d.GetOk("kandji"); ok {
		kandjiConfigs := v.(*schema.Set).List()
		if len(kandjiConfigs) > 0 {
			kandjiConfig := kandjiConfigs[0].(map[string]interface{})

			request := corev1.CreateIntegrationMDMRequest{
				Name: d.Get("name").(string),
				Integration: &corev1.CreateIntegrationMDMRequest_Kandji_{
					Kandji: &corev1.CreateIntegrationMDMRequest_Kandji{
						ApiKey: kandjiConfig["api_key"].(string),
						ApiUrl: kandjiConfig["api_url"].(string),
					},
				},
			}
			res, err = c.Grpc.Sdk.IntegrationMDMServiceClient.CreateIntegrationMDM(ctx, connect.NewRequest(&request))
		}
	}

	// Handle error if any
	if err != nil {
		return diag.FromErr(err)
	}

	// Assuming you need to handle a situation where none are configured
	if res == nil {
		return diag.Errorf("No integration configuration found")
	}

	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationMDMRead(ctx, d, m)
	return diags
}

func resourceIntegrationMDMRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	res, err := c.Grpc.Sdk.IntegrationMDMServiceClient.GetIntegrationMDM(ctx, connect.NewRequest(&corev1.GetIntegrationMDMRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Integration.Name)
	d.Set("integration", res.Msg.Integration.Integration)
	d.Set("created_at", res.Msg.Integration.CreatedAt.AsTime().Unix())

	d.SetId(res.Msg.Integration.Id)

	return diags
}

func resourceIntegrationMDMDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	_, err := c.Grpc.Sdk.IntegrationMDMServiceClient.DeleteIntegrationMDM(ctx, connect.NewRequest(&corev1.DeleteIntegrationMDMRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
