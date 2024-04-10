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

func ResourcePolicyExternalDataLoader() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a External Data Loader External Api.",
		CreateContext: ResourcePolicyExternalDataLoaderCreate,
		ReadContext:   ResourcePolicyExternalDataLoaderRead,
		DeleteContext: ResourcePolicyExternalDataLoaderDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Policy External Data Loader.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the Policy External Data Loader.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"host": {
				// This description is used by the documentation generator and the language server.
				Description: "Host of the Policy External Data Loader.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "Port of the ExternalApi: basic or oauth",
				Type:        schema.TypeInt,
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

func ResourcePolicyExternalDataLoaderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	host := d.Get("host").(string)
	authType := d.Get("auth_type").(string)
	basicAuthUsername := d.Get("basic_auth_username").(string)
	basicAuthPassword := d.Get("basic_auth_password").(string)
	port := d.Get("port").(int)

	externalApi, err := c.Grpc.Sdk.PolicyServiceClient.CreateExternalDataLoader(ctx, connect.NewRequest(&corev1.CreateExternalDataLoaderRequest{
		Name: name,
		Host: host,
		Port: int32(port),
		Auth: &corev1.Auth{
			Type: authType,
			Basic: &corev1.Auth_Basic{
				Username: basicAuthUsername,
				Password: basicAuthPassword,
			},
		},
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(externalApi.Msg.ExternalDataLoader.Id)

	ResourcePolicyExternalDataLoaderRead(ctx, d, meta)
	return diags
}

func ResourcePolicyExternalDataLoaderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	externalApi, err := c.Grpc.Sdk.PolicyServiceClient.GetExternalDataLoader(ctx, connect.NewRequest(&corev1.GetExternalDataLoaderRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", externalApi.Msg.ExternalDataLoader.Name)
	d.Set("host", externalApi.Msg.ExternalDataLoader.Host)
	d.Set("port", externalApi.Msg.ExternalDataLoader.Port)

	d.SetId(externalApi.Msg.ExternalDataLoader.Id)

	return diags
}

func ResourcePolicyExternalDataLoaderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.PolicyServiceClient.DeleteExternalDataLoader(ctx, connect.NewRequest(&corev1.DeleteExternalDataLoaderRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
