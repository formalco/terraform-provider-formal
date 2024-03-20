package resource

import (
	"context"
	"time"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIntegrationExternalApi() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Integration External Api.",
		CreateContext: resourceIntegrationExternalApiCreate,
		ReadContext:   resourceIntegrationExternalApiRead,
		DeleteContext: resourceIntegrationExternalApiDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the ExternalApi.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the ExternalApi.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the ExternalApi: zendesk or custom",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				// This description is used by the documentation generator and the language server.
				Description: "Url of the ExternalApi.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"auth_type": {
				// This description is used by the documentation generator and the language server.
				Description: "Auth type of the ExternalApi: basic or oauth",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"basic_auth_username": {
				// This description is used by the documentation generator and the language server.
				Description: "Username of the ExternalApi.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"basic_auth_password": {
				// This description is used by the documentation generator and the language server.
				Description: "Password of the ExternalApi.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIntegrationExternalApiCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	externapApiType := d.Get("type").(string)
	authType := d.Get("auth_type").(string)
	basicAuthUsername := d.Get("basic_auth_username").(string)
	basicAuthPassword := d.Get("basic_auth_password").(string)
	url := d.Get("url").(string)

	externalApi, err := c.Grpc.Sdk.ExternalApiServiceClient.CreateExternalApiIntegration(ctx, connect.NewRequest(&adminv1.CreateExternalApiIntegrationRequest{
		Type: externapApiType,
		Name: name,
		Url:  url,
		Auth: &adminv1.Auth{
			Type: authType,
			Basic: &adminv1.Auth_Basic{
				Username: basicAuthUsername,
				Password: basicAuthPassword,
			},
		},
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(externalApi.Msg.Id)

	resourceIntegrationExternalApiRead(ctx, d, meta)
	return diags
}

func resourceIntegrationExternalApiRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	externalApi, err := c.Grpc.Sdk.ExternalApiServiceClient.GetExternalApiIntegration(ctx, connect.NewRequest(&adminv1.GetExternalApiIntegrationRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", externalApi.Msg.Integration.Name)
	d.Set("type", externalApi.Msg.Integration.Type)
	d.Set("url", externalApi.Msg.Integration.Url)
	d.Set("auth_type", externalApi.Msg.Integration.AuthType)

	d.SetId(externalApi.Msg.Integration.Id)

	return diags
}

func resourceIntegrationExternalApiDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.ExternalApiServiceClient.DeleteExternalApiIntegration(ctx, connect.NewRequest(&adminv1.DeleteExternalApiIntegrationRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
