package resource

import (
	"context"
	"fmt"
	"strings"
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
				Description: "Auth type of the ExternalApi: basic or oauth",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					validTypes := []string{"basic", "oauth"}
					for _, t := range validTypes {
						if v == t {
							return
						}
					}
					errs = append(errs, fmt.Errorf("%q must be one of [%s], got %q", key, strings.Join(validTypes, ", "), v))
					return
				},
			},
			"basic_auth": {
				Type:        schema.TypeSet, // Use TypeList or TypeSet based on whether the order matters or uniqueness is required
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1, // Ensures only one object is provided
				Description: "Basic authentication credentials.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username of the ExternalApi.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"password": {
							Description: "Password of the ExternalApi.",
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

func ResourcePolicyExternalDataLoaderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	host := d.Get("host").(string)
	authType := d.Get("auth_type").(string)
	port := d.Get("port").(int)

	var basicAuth *corev1.Auth_Basic
	if v, ok := d.GetOk("basic_auth"); ok {
		basicAuthList := v.(*schema.Set).List()
		if len(basicAuthList) > 0 {
			basicAuthMap := basicAuthList[0].(map[string]interface{})
			basicAuthUsername := basicAuthMap["username"].(string)
			basicAuthPassword := basicAuthMap["password"].(string)
			basicAuth = &corev1.Auth_Basic{
				Username: basicAuthUsername,
				Password: basicAuthPassword,
			}
		}
	}

	externalApi, err := c.Grpc.Sdk.PolicyServiceClient.CreateExternalDataLoader(ctx, connect.NewRequest(&corev1.CreateExternalDataLoaderRequest{
		Name: name,
		Host: host,
		Port: int32(port),
		Auth: &corev1.Auth{
			Type:  authType,
			Basic: basicAuth,
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
