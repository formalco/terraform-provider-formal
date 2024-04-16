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
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Name of the Integration",
				Type:        schema.TypeString,
				Required:    true,
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
			"duo": {
				Description: "Configuration block for Duo integration. This block is optional and may be omitted if not configuring a Duo integration.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_key": {
							Description: "Duo Integration Key.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"secret_key": {
							Description: "Duo Secret Key.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
						"api_hostname": {
							Description: "Duo API Hostname.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
		},
	}
}

func resourceIntegrationMfaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	var res *connect.Response[corev1.CreateIntegrationMfaResponse]
	var err error

	// Check if the 'duo' configuration block is present
	if v, ok := d.GetOk("duo"); ok {
		duoConfigs := v.(*schema.Set).List()
		if len(duoConfigs) > 0 {
			// Assuming there's only one 'duo' configuration block
			duoConfig := duoConfigs[0].(map[string]interface{})

			duo := &corev1.CreateIntegrationMfaRequest_Duo_{
				Duo: &corev1.CreateIntegrationMfaRequest_Duo{
					IntegrationKey: duoConfig["integration_key"].(string),
					SecretKey:      duoConfig["secret_key"].(string),
					ApiHostname:    duoConfig["api_hostname"].(string),
				},
			}
			res, err = c.Grpc.Sdk.IntegrationMfaServiceClient.CreateIntegrationMfa(ctx, connect.NewRequest(&corev1.CreateIntegrationMfaRequest{
				Name: d.Get("name").(string),
				Mfa:  duo,
			}))
		} else {
			return diag.Errorf("Duo configuration is required.")
		}
	} else {
		return diag.Errorf("No MFA configuration found.")
	}

	if err != nil {
		return diag.FromErr(err)
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

	resSecrets, err := c.Grpc.Sdk.IntegrationMfaServiceClient.GetIntegrationMFASecrets(ctx, connect.NewRequest(&corev1.GetIntegrationMFASecretsRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if data, ok := res.Msg.Integration.Mfa.(*corev1.IntegrationMfa_Duo_); ok {
		// Construct a map for the 'duo' configuration
		duoConfig := map[string]interface{}{
			"api_hostname": data.Duo.ApiHostname,
		}

		if secrets, ok := resSecrets.Msg.Mfa.(*corev1.GetIntegrationMFASecretsResponse_Duo_); ok {
			duoConfig["secret_key"] = secrets.Duo.SecretKey
			duoConfig["integration_key"] = secrets.Duo.IntegrationKey
		}

		// Create a new set for the 'duo' configuration
		duoSet := schema.NewSet(schema.HashResource(ResourceIntegrationMfa()), []interface{}{duoConfig})
		if err := d.Set("duo", duoSet); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// If not a Duo MFA type or not set, ensure the 'duo' field in Terraform state is cleared or handled as needed
		d.Set("duo", nil)
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
