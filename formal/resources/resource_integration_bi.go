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
			"sync": {
				// This description is used by the documentation generator and the language server.
				Description: "Auto synchronize users from Metabase to Formal (occurs every hour). Note that a lambda worker will need to be deployed in your infrastructure to synchronise users.",
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
			},
			"metabase": {
				Description: "Configuration block for Metabase integration. This block is optional and may be omitted if not configuring a Metabase integration.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Description: "Hostname of the Metabase instance.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"username": {
							Description: "Username for the Metabase instance.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"password": {
							Description: "Password for the Metabase instance.",
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

func resourceIntegrationBICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	Name := d.Get("name").(string)
	Sync := d.Get("sync").(bool)

	var res *connect.Response[corev1.CreateBIIntegrationResponse]
	var err error

	if !Sync {
		res, err = c.Grpc.Sdk.IntegrationBIServiceClient.CreateBIIntegration(ctx, connect.NewRequest(&corev1.CreateBIIntegrationRequest{
			Name: Name,
			Sync: Sync,
		}))
		if err != nil {
			return diag.FromErr(err)
		}

	} else {
		if v, ok := d.GetOk("metabase"); ok {
			metabaseSet := v.(*schema.Set)

			// As we expect only one item in the set, we can iterate through the set
			// Though not typical for sets, this is a pattern in Terraform when a set is used to ensure uniqueness
			for _, metabaseConfig := range metabaseSet.List() {
				config := metabaseConfig.(map[string]interface{})

				metabase := &corev1.CreateBIIntegrationRequest_Metabase_{
					Metabase: &corev1.CreateBIIntegrationRequest_Metabase{
						Hostname: config["hostname"].(string),
						Username: config["username"].(string),
						Password: config["password"].(string),
					},
				}
				res, err = c.Grpc.Sdk.IntegrationBIServiceClient.CreateBIIntegration(ctx, connect.NewRequest(&corev1.CreateBIIntegrationRequest{
					Name: Name,
					Sync: Sync,
					Type: metabase,
				}))
				if err != nil {
					return diag.FromErr(err)
				}
				// Proceed with using the 'metabase' variable for your request
				// Since we expect only one metabase configuration block,
				// we can break the loop after processing the first item
				break
			}
		} else {
			return diag.Errorf("Unsupported bi type")
		}

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
	d.Set("sync", res.Msg.Integration.Sync)

	switch data := res.Msg.Integration.Type.(type) {
	case *corev1.BIIntegration_Metabase_:
		metabaseConfig := map[string]interface{}{
			"hostname": data.Metabase.Hostname,
			"username": data.Metabase.Username,
			"password": data.Metabase.Password,
		}

		// Create a new set to store the metabase configuration
		metabaseSet := schema.NewSet(schema.HashResource(ResourceIntegrationBI()), []interface{}{metabaseConfig})
		d.Set("metabase", metabaseSet)
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
