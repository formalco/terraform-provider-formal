package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
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
				Description: "Friendly name for this app.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"sync": {
				// This description is used by the documentation generator and the language server.
				Description: "Auto synchronize users from Metabase to Formal (occurs every hour). When disabled, a worker will need to be deployed in your infrastructure to synchronise users.",
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
							Description: "Metabase server hostname. Required when `sync=true`.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"username": {
							Description: "Metabase admin username. Required when `sync=true`.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"password": {
							Description: "Metabase admin password. Required when `sync=true`.",
							Type:        schema.TypeString,
							Optional:    true,
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

	biIntegration := &corev1.CreateBIIntegrationRequest{
		Name: d.Get("name").(string),
		Sync: d.Get("sync").(bool),
	}

	if metabaseRaw, ok := d.GetOk("metabase"); ok {
		metabaseSet := metabaseRaw.(*schema.Set)

		// As we expect only one item in the set, we can iterate through the set
		// Though not typical for sets, this is a pattern in Terraform when a set is used to ensure uniqueness
		for _, metabaseConfig := range metabaseSet.List() {
			config := metabaseConfig.(map[string]interface{})

			metabase := &corev1.CreateBIIntegrationRequest_Metabase{}
			if val, exists := config["hostname"]; exists && val != nil {
				metabase.Hostname = val.(string)
			}
			if val, exists := config["username"]; exists && val != nil {
				metabase.Username = val.(string)
			}
			if val, exists := config["password"]; exists && val != nil {
				metabase.Password = val.(string)
			}

			if biIntegration.Sync {
				if metabase.Hostname == "" {
					return diag.Errorf("metabase hostname is required when sync=true")
				}
				if metabase.Username == "" {
					return diag.Errorf("metabase username is required when sync=true")
				}
				if metabase.Password == "" {
					return diag.Errorf("metabase password is required when sync=true")
				}
			}

			biIntegration.Type = &corev1.CreateBIIntegrationRequest_Metabase_{
				Metabase: metabase,
			}
			break
		}
	}

	res, err := c.Grpc.Sdk.IntegrationBIServiceClient.CreateBIIntegration(ctx, connect.NewRequest(biIntegration))
	if err != nil {
		return diag.FromErr(err)
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
	d.Set("metabase", nil)

	switch data := res.Msg.Integration.Type.(type) {
	case *corev1.BIIntegration_Metabase_:
		metabaseConfig := map[string]interface{}{
			"hostname": data.Metabase.Hostname,
			"username": data.Metabase.Username,
			"password": data.Metabase.Password,
		}

		// Create a new set to store the metabase configuration
		metabaseSet := schema.NewSet(schema.HashResource(ResourceIntegrationBI().Schema["metabase"].Elem.(*schema.Resource)), []interface{}{metabaseConfig})
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
