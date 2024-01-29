package resource

import (
	"context"
	"time"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"

	"github.com/bufbuild/connect-go"
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
		UpdateContext: resourceIntegrationMfaUpdate,
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
			},
		},
	}
}

func resourceIntegrationMfaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	typeApp := d.Get("type").(string)

	duoIntegrationKey := d.Get("duo_integration_key").(string)
	duoSecretKey := d.Get("duo_secret_key").(string)
	duoApiHostname := d.Get("duo_api_hostname").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.IntegrationMfaServiceClient.CreateIntegrationMfa(ctx, connect.NewRequest(&adminv1.CreateIntegrationMfaRequest{
		Type:                  typeApp,
		DuoIntegrationKey:     duoIntegrationKey,
		DuoSecretKey:          duoSecretKey,
		DuoApiHostname:        duoApiHostname,
		TerminationProtection: terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)

	resourceIntegrationMfaRead(ctx, d, m)
	return diags
}

func resourceIntegrationMfaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	res, err := c.Grpc.Sdk.IntegrationMfaServiceClient.GetIntegrationMfaById(ctx, connect.NewRequest(&adminv1.GetIntegrationMfaByIdRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("type", res.Msg.Integration.Type)
	d.Set("duo_integration_key", res.Msg.Integration.DuoIntegrationKey)
	d.Set("duo_secret_key", res.Msg.Integration.DuoSecretKey)
	d.Set("duo_api_hostname", res.Msg.Integration.DuoApiHostname)
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

	_, err := c.Grpc.Sdk.IntegrationMfaServiceClient.DeleteIntegrationMfa(ctx, connect.NewRequest(&adminv1.DeleteIntegrationMfaRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceIntegrationMfaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.IntegrationMfaServiceClient.UpdateIntegrationMfa(ctx, connect.NewRequest(&adminv1.UpdateIntegrationMfaRequest{
			Id:                    d.Id(),
			TerminationProtection: terminationProtection,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceIntegrationMfaRead(ctx, d, meta)

	return diags
}
