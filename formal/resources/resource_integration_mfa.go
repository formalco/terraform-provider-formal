package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"

	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIntegrationMfa() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Integration MFA app.",
		CreateContext: resourceIntegrationMfaCreate,
		ReadContext:   resourceIntegrationMfaRead,
		DeleteContext: resourceIntegrationMfaDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Integration Mfa.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the Integration mfa app: `duo`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"duo_integration_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Duo Integration Key.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"duo_secret_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Duo Secret Key.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"duo_api_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Duo API Hostname.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Integration MFA cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
		},
	}
}

func resourceIntegrationMfaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	typeApp := d.Get("type").(string)

	var res *connect.Response[corev1.CreateIntegrationMfaResponse]
	var err error

	switch typeApp {
	case "duo":
		duo := &corev1.CreateIntegrationMfaRequest_Duo_{
			Duo: &corev1.CreateIntegrationMfaRequest_Duo{
				IntegrationKey: d.Get("duo_integration_key").(string),
				SecretKey:      d.Get("duo_secret_key").(string),
				ApiHostname:    d.Get("duo_api_hostname").(string),
			},
		}
		res, err = c.Grpc.Sdk.IntegrationMfaServiceClient.CreateIntegrationMfa(ctx, connect.NewRequest(&corev1.CreateIntegrationMfaRequest{
			Name: d.Get("name").(string),
			Mfa:  duo,
		}))
		if err != nil {
			return diag.FromErr(err)
		}

	default:
		return diag.Errorf("Unsupported mfa type: %s", typeApp)
	}

	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationMfaRead(ctx, d, m)
	return diags
}

func resourceIntegrationMfaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	res, err := c.Grpc.Sdk.IntegrationMfaServiceClient.GetIntegrationMfa(ctx, connect.NewRequest(&corev1.GetIntegrationMfaRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	switch data := res.Msg.Integration.Mfa.(type) {
	case *corev1.IntegrationMfa_Duo_:
		d.Set("duo_api_hostname", data.Duo.ApiHostname)
	}

	d.Set("termination_protection", res.Msg.Integration.TerminationProtection)

	d.SetId(res.Msg.Integration.Id)

	return diags
}

func resourceIntegrationMfaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Integration MFA cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.IntegrationMfaServiceClient.DeleteIntegrationMfa(ctx, connect.NewRequest(&corev1.DeleteIntegrationMfaRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
