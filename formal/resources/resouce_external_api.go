package resource

import (
	"context"
	"github.com/formalco/terraform-provider-formal/formal/apiv2"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceExternalApi() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a External Api in Formal.",

		CreateContext: resourceExternalApiCreate,
		ReadContext:   resourceExternalApiRead,
		//UpdateContext: resourceExternalApiUpdate,
		DeleteContext: resourceExternalApiDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for this external api.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of external api.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "The name of the external api.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"url": {
				// This description is used by the documentation generator and the language server.
				Description: "The url of the external api.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"auth_username": {
				// This description is used by the documentation generator and the language server.
				Description: "The username for authentication.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"auth_password": {
				// This description is used by the documentation generator and the language server.
				Description: "The password for authentication.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"keyword": {
				// This description is used by the documentation generator and the language server.
				Description: "The keyword for authentication.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// Done
func resourceExternalApiCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	newExternalApi := apiv2.IntegrationExternalAPI{
		Type:    d.Get("type").(string),
		Name:    d.Get("name").(string),
		Url:     d.Get("url").(string),
		Keyword: d.Get("keyword").(string),
		Auth: apiv2.IntegrationExternalAPIAuth{
			Type: "basic",
			Basic: apiv2.IntegrationExternalAPIAuthBasic{
				Username: d.Get("auth_username").(string),
				Password: d.Get("auth_password").(string),
			},
		},
	}

	externalApiId, err := c.Grpc.CreateExternalAPIIntegration(ctx, newExternalApi)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(externalApiId)

	resourceExternalApiRead(ctx, d, meta)

	return diags
}

func resourceExternalApiRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	externalAPIIntegration, err := c.Grpc.GetExternalAPIIntegration(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("type", externalAPIIntegration.Type)
	d.Set("name", externalAPIIntegration.Name)
	d.Set("url", externalAPIIntegration.Url)
	d.Set("keyword", externalAPIIntegration.Keyword)
	d.Set("auth_username", externalAPIIntegration.Auth.Basic.Username)
	d.Set("auth_password", externalAPIIntegration.Auth.Basic.Password)

	d.SetId(externalAPIIntegration.ID)

	return diags
}

func resourceExternalApiDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	externalApiId := d.Id()

	err := c.Grpc.DeleteExternalAPIIntegration(ctx, externalApiId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
