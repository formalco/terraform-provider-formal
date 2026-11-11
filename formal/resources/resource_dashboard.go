package resource

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceDashboard() *schema.Resource {
	return &schema.Resource{
		Description:   "A saved dashboard of logs, AI usage, and note widgets.",
		CreateContext: resourceDashboardCreate,
		ReadContext:   resourceDashboardRead,
		UpdateContext: resourceDashboardUpdate,
		DeleteContext: resourceDashboardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the dashboard.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the dashboard.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"definition": {
				Description:      "The JSON form of the `core.v1.DashboardDefinition` message, usually written with `jsonencode`.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentDashboardDefinition,
			},
		},
	}
}

func suppressEquivalentDashboardDefinition(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	oldDefinition, newDefinition := &corev1.DashboardDefinition{}, &corev1.DashboardDefinition{}
	return protojson.Unmarshal([]byte(oldValue), oldDefinition) == nil &&
		protojson.Unmarshal([]byte(newValue), newDefinition) == nil &&
		proto.Equal(oldDefinition, newDefinition)
}

func dashboardDefinition(d *schema.ResourceData) (*corev1.DashboardDefinition, error) {
	definition := &corev1.DashboardDefinition{}
	return definition, protojson.Unmarshal([]byte(d.Get("definition").(string)), definition)
}

func flattenDashboard(d *schema.ResourceData, dashboard *corev1.Dashboard) diag.Diagnostics {
	definition, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(dashboard.Definition)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(dashboard.Id)
	_ = d.Set("name", dashboard.Name)
	_ = d.Set("description", dashboard.Description)
	_ = d.Set("definition", string(definition))
	return nil
}

func resourceDashboardCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	definition, err := dashboardDefinition(d)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := c.Grpc.Sdk.DashboardServiceClient.CreateDashboard(ctx, &corev1.CreateDashboardRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Definition:  definition,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return flattenDashboard(d, res.Dashboard)
}

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	res, err := c.Grpc.Sdk.DashboardServiceClient.GetDashboard(ctx, &corev1.GetDashboardRequest{Id: d.Id()})
	if connect.CodeOf(err) == connect.CodeNotFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return flattenDashboard(d, res.Dashboard)
}

func resourceDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	definition, err := dashboardDefinition(d)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := c.Grpc.Sdk.DashboardServiceClient.UpdateDashboard(ctx, &corev1.UpdateDashboardRequest{
		Id:          d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Definition:  definition,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return flattenDashboard(d, res.Dashboard)
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	_, err := c.Grpc.Sdk.DashboardServiceClient.DeleteDashboard(ctx, &corev1.DeleteDashboardRequest{Id: d.Id()})
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return diag.FromErr(err)
	}
	return nil
}
