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

func ResourceIntegrationBI() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a BI App.",
		CreateContext: resourceIntegrationBICreate,
		ReadContext:   resourceIntegrationBIRead,
		DeleteContext: resourceIntegrationBIDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the App.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the App.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the App: `metabase` or `fivetran`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"metabase_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Hostname of the Metabase app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"metabase_username": {
				// This description is used by the documentation generator and the language server.
				Description: "Username of the Metabase app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"metabase_password": {
				// This description is used by the documentation generator and the language server.
				Description: "Password of the Metabase app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"fivetran_api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "API Key of Fivetran app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"fivetran_api_secret": {
				// This description is used by the documentation generator and the language server.
				Description: "Secret of Fivetran app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"linked_db_user_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Linked DB User ID of the App.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIntegrationBICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	Name := d.Get("name").(string)
	LinkedDBUserID := d.Get("linked_db_user_id").(string)

	typeApp := d.Get("type").(string)

	var res *connect.Response[corev1.CreateBIIntegrationResponse]
	var err error

	switch typeApp {
	case "metabase":
		metabase := &corev1.CreateBIIntegrationRequest_Metabase_{
			Metabase: &corev1.CreateBIIntegrationRequest_Metabase{
				Hostname: d.Get("metabase_hostname").(string),
				Username: d.Get("metabase_username").(string),
				Password: d.Get("metabase_password").(string),
			},
		}
		res, err = c.Grpc.Sdk.IntegrationBIServiceClient.CreateBIIntegration(ctx, connect.NewRequest(&corev1.CreateBIIntegrationRequest{
			Name:           Name,
			Type:           metabase,
			LinkedDbUserId: LinkedDBUserID,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	case "fivetran":
		fivetran := &corev1.CreateBIIntegrationRequest_Fivetran_{
			Fivetran: &corev1.CreateBIIntegrationRequest_Fivetran{
				FivetranApiKey:    d.Get("fivetran_api_key").(string),
				FivetranApiSecret: d.Get("fivetran_api_secret").(string),
			},
		}
		res, err = c.Grpc.Sdk.IntegrationBIServiceClient.CreateBIIntegration(ctx, connect.NewRequest(&corev1.CreateBIIntegrationRequest{
			Name:           Name,
			Type:           fivetran,
			LinkedDbUserId: LinkedDBUserID,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Unsupported mfa type: %s", typeApp)
	}

	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationBIRead(ctx, d, meta)
	return diags
}

func resourceIntegrationBIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	res, err := c.Grpc.Sdk.IntegrationBIServiceClient.GetBIIntegration(ctx, connect.NewRequest(&corev1.GetBIIntegrationRequest{
		Id: appId,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Integration.Name)
	d.Set("type", res.Msg.Integration.Type)
	d.Set("linked_db_user_id", res.Msg.Integration.LinkedDbUserId)

	switch data := res.Msg.Integration.Type.(type) {
	case *corev1.BIIntegration_Metabase_:
		d.Set("metabase_hostname", data.Metabase.Hostname)
		d.Set("metabase_username", data.Metabase.Username)
		d.Set("metabase_password", data.Metabase.Password)
	case *corev1.BIIntegration_Fivetran_:
		d.Set("fivetran_api_key", data.Fivetran.ApiKey)
		d.Set("fivetran_api_secret", data.Fivetran.ApiSecret)
	}

	d.SetId(res.Msg.Integration.Id)
	return diags
}

func resourceIntegrationBIDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	_, err := c.Grpc.Sdk.IntegrationBIServiceClient.DeleteBIIntegration(ctx, connect.NewRequest(&corev1.DeleteBIIntegrationRequest{Id: appId}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
