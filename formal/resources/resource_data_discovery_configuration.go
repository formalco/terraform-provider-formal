package resource

import (
	"context"
	"net/http"

	core_connect "buf.build/gen/go/formal/core/connectrpc/go/core/v1/corev1connect"
	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type transport struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Api-Key", t.apiKey)
	return t.underlyingTransport.RoundTrip(req)
}

func newApiClientV2(apiKey string) *http.Client {
	httpClient := &http.Client{Transport: &transport{
		underlyingTransport: http.DefaultTransport,
		apiKey:              apiKey,
	}}
	return httpClient
}

func ResourceDataDiscoveryConfiguration() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Data Discovery Configuration in Formal.",

		CreateContext: resourceDataDiscoveryConfigurationCreate,
		ReadContext:   resourceDataDiscoveryConfigurationRead,
		UpdateContext: resourceDataDiscoveryConfigurationUpdate,
		DeleteContext: resourceDataDiscoveryConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for this Data Discovery Configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The resource scanned by this Data Discovery Configuration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"native_user_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The native user used to scan the resource.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"discovery_schedule": {
				// This description is used by the documentation generator and the language server.
				Description: "The schedule at which the discovery is being run.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"deletion_policy": {
				// This description is used by the documentation generator and the language server.
				Description: "The deletion policy for the Data Discovery Configuration. It can be either: `delete` or `mark_for_deletion`.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceDataDiscoveryConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	resourceId := d.Get("resource_id").(string)
	nativeUserId := d.Get("native_user_id").(string)
	discoverySchedule := d.Get("discovery_schedule").(string)
	deletionPolicy := d.Get("deletion_policy").(string)

	httpClient := newApiClientV2(c.ApiKey)
	client := core_connect.NewResourceServiceClient(httpClient, "https://v2api.formalcloud.net")
	res, err := client.CreateDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.CreateDataDiscoveryConfigurationRequest{
		ResourceId:     resourceId,
		NativeUserId:   nativeUserId,
		Schedule:       discoverySchedule,
		DeletionPolicy: deletionPolicy,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.DataDiscoveryConfiguration.Id)

	resourceDataDiscoveryConfigurationRead(ctx, d, meta)
	return diags
}

func resourceDataDiscoveryConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	id := d.Id()

	httpClient := newApiClientV2(c.ApiKey)
	client := core_connect.NewResourceServiceClient(httpClient, "https://v2api.formalcloud.net")
	res, err := client.GetDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.GetDataDiscoveryConfigurationRequest{
		Id: &corev1.GetDataDiscoveryConfigurationRequest_DataDiscoveryConfigurationId{
			DataDiscoveryConfigurationId: id,
		}}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Group was deleted
			tflog.Warn(ctx, "The Data Discovery Configuration was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of
	d.Set("id", res.Msg.DataDiscoveryConfiguration.Id)
	d.Set("resource_id", res.Msg.DataDiscoveryConfiguration.ResourceId)
	d.Set("native_user_id", res.Msg.DataDiscoveryConfiguration.NativeUserId)
	d.Set("discovery_schedule", res.Msg.DataDiscoveryConfiguration.Schedule)
	d.Set("deletion_policy", res.Msg.DataDiscoveryConfiguration.DeletionPolicy)

	d.SetId(id)

	return diags
}

func resourceDataDiscoveryConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()
	nativeUserId := d.Get("native_user_id").(string)
	discoverySchedule := d.Get("discovery_schedule").(string)
	deletionPolicy := d.Get("deletion_policy").(string)

	httpClient := newApiClientV2(c.ApiKey)
	client := core_connect.NewResourceServiceClient(httpClient, "https://v2api.formalcloud.net")
	_, err := client.UpdateDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.UpdateDataDiscoveryConfigurationRequest{
		Id:             id,
		NativeUserId:   nativeUserId,
		Schedule:       discoverySchedule,
		DeletionPolicy: deletionPolicy,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Group was deleted
			tflog.Warn(ctx, "The Data Discovery Configuration was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	resourceGroupRead(ctx, d, meta)

	return diags
}

func resourceDataDiscoveryConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Group cannot be deleted because termination_protection is set to true")
	}

	id := d.Id()

	httpClient := newApiClientV2(c.ApiKey)
	client := core_connect.NewResourceServiceClient(httpClient, "https://v2api.formalcloud.net")
	_, err := client.DeleteDataDiscoveryConfiguration(ctx, connect.NewRequest(&corev1.DeleteDataDiscoveryConfigurationRequest{Id: id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
