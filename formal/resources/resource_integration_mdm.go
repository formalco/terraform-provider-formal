package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceIntegrationMDM() *schema.Resource {
	return &schema.Resource{
		Description:   "Registering a Integration MDM app.",
		CreateContext: resourceIntegrationMDMCreate,
		ReadContext:   resourceIntegrationMDMRead,
		UpdateContext: resourceIntegrationMDMUpdate,
		DeleteContext: resourceIntegrationMDMDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourceIntegrationMDMV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateIntegrationMDMStateV0,
			},
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
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {
							Description: "API Key of your Kandji organization. This value is not stored in Terraform state. To rotate the key, change this value and run `terraform apply -replace=<resource address>`.",
							Type:        schema.TypeString,
							Required:    true,
							WriteOnly:   true,
						},
						"api_url": {
							Description: "API URL of your Kandji organization.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceIntegrationMDMCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*clients.Clients)

	kandjiList, ok := d.GetOk("kandji")
	if !ok || len(kandjiList.([]any)) == 0 {
		return diag.Errorf("kandji integration configuration is required")
	}

	kandjiConfig := kandjiList.([]any)[0].(map[string]any)

	apiKeyVal, rawDiags := d.GetRawConfigAt(cty.GetAttrPath("kandji").Index(cty.NumberIntVal(0)).GetAttr("api_key"))
	if rawDiags.HasError() {
		return diag.Errorf("failed to get kandji api_key: %v", rawDiags)
	}
	if apiKeyVal.IsNull() || apiKeyVal.Type() != cty.String {
		return diag.Errorf("kandji api_key is required")
	}

	res, err := c.Grpc.Sdk.IntegrationMDMServiceClient.CreateIntegrationMDM(ctx, connect.NewRequest(&corev1.CreateIntegrationMDMRequest{
		Name: d.Get("name").(string),
		Integration: &corev1.CreateIntegrationMDMRequest_Kandji_{
			Kandji: &corev1.CreateIntegrationMDMRequest_Kandji{
				ApiKey: apiKeyVal.AsString(),
				ApiUrl: kandjiConfig["api_url"].(string),
			},
		},
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Integration.Id)

	return resourceIntegrationMDMRead(ctx, d, m)
}

func resourceIntegrationMDMRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if kandji := res.Msg.Integration.GetKandji(); kandji != nil {
		d.Set("kandji", []map[string]any{
			{"api_url": kandji.GetApiUrl()},
		})
	}

	d.SetId(res.Msg.Integration.Id)

	return diags
}

func resourceIntegrationMDMUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return diag.Errorf("api_key cannot be updated in-place; use `terraform apply -replace=<resource address>` to recreate the resource (for example, `terraform apply -replace=formal_integration_mdm.example`)")
}

func resourceIntegrationMDMV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id":   {Type: schema.TypeString, Computed: true},
			"name": {Type: schema.TypeString, Required: true},
			"kandji": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {Type: schema.TypeString, Required: true},
						"api_url": {Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	}
}

func migrateIntegrationMDMStateV0(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	elems, ok := kandjiElementsFromV0State(rawState["kandji"])
	if !ok {
		return rawState, nil
	}

	migrated := make([]any, 0, len(elems))
	for _, elem := range elems {
		migrated = append(migrated, map[string]any{"api_url": elem["api_url"]})
	}
	rawState["kandji"] = migrated

	return rawState, nil
}

func kandjiElementsFromV0State(raw any) ([]map[string]any, bool) {
	var items []any

	switch v := raw.(type) {
	case *schema.Set:
		if v.Len() == 0 {
			return nil, false
		}
		items = v.List()
	case []any:
		items = v
	default:
		return nil, false
	}

	elems := make([]map[string]any, 0, len(items))
	for _, item := range items {
		elem, ok := item.(map[string]any)
		if !ok {
			return nil, false
		}
		elems = append(elems, elem)
	}

	return elems, len(elems) > 0
}

func resourceIntegrationMDMDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
