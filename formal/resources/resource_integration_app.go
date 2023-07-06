package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func ResourceIntegrationApp() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Integration App.",
		CreateContext: resourceIntegrationAppCreate,
		ReadContext:   resourceIntegrationAppRead,
		DeleteContext: resourceIntegrationAppDelete,

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
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the App: metabase or custom",
				Type:        schema.TypeString,
				Required:    true,
			},
			"metabase_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "Hostname of the Metabase app.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"metabase_username": {
				// This description is used by the documentation generator and the language server.
				Description: "Username of the Metabase app.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"metabase_password": {
				// This description is used by the documentation generator and the language server.
				Description: "Password of the Metabase app.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"linked_db_user_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Linked DB User ID of the App.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceIntegrationAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	Name := d.Get("name").(string)
	Type := d.Get("type").(string)
	MetabaseHostname := d.Get("metabase_hostname").(string)
	MetabaseUsername := d.Get("metabase_username").(string)
	MetabasePassword := d.Get("metabase_password").(string)
	LinkedDBUserID := d.Get("linked_db_user_id").(string)

	res, err := c.Grpc.Sdk.AppServiceClient.CreateIntegrationApp(ctx, connect.NewRequest(&adminv1.CreateIntegrationAppRequest{
		Name:             Name,
		Type:             Type,
		MetabaseHostname: MetabaseHostname,
		MetabaseUsername: MetabaseUsername,
		MetabasePassword: MetabasePassword,
		LinkedDbUserId:   LinkedDBUserID,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationAppRead(ctx, d, meta)
	return diags
}

func resourceIntegrationAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//c := meta.(*clients.Clients)
	var diags diag.Diagnostics
	//
	//appId := d.Id()
	//
	//res, err := c.Grpc.Sdk.AppServiceClient.GetIntegrationApps(ctx, connect.NewRequest(&adminv1.GetIntegrationApps{}))
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	return diags
}

func resourceIntegrationAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	_, err := c.Grpc.Sdk.AppServiceClient.DeleteIntegrationApp(ctx, connect.NewRequest(&adminv1.DeleteIntegrationAppRequest{Id: appId}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
