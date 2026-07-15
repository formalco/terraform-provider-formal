package resource

import (
	"context"
	"fmt"
	"regexp"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceHook() *schema.Resource {
	return &schema.Resource{
		Description: "Hooks are JavaScript functions evaluated during policy decisions. Policies reference hooks as `input.hooks.<name>`.",

		CreateContext: resourceHookCreate,
		ReadContext:   resourceHookRead,
		UpdateContext: resourceHookUpdate,
		DeleteContext: resourceHookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the hook.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description:  "The name of the hook. Must be unique within the organization and match `^[A-Za-z_][A-Za-z0-9_]*$`. Policies reference this name as `input.hooks.<name>`.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`), "name must match ^[A-Za-z_][A-Za-z0-9_]*$"),
			},
			"description": {
				Description: "The hook description.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"code": {
				Description: "The hook implementation as JavaScript. Must be a default-exported function (for example `export default function hook(input, env) { ... }`). The optional second argument receives allowlisted process environment variables.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "The hook status. Accepted values are `active` and `draft`. Only active hooks can be referenced by policies.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "draft",
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"draft",
				}, false),
			},
			"timeout_ms": {
				Description:  "Maximum time in milliseconds the hook may run during policy evaluation. Must be between 1 and 60000.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5000,
				ValidateFunc: validation.IntBetween(1, 60000),
			},
			"allowlisted_environment_variables": {
				Description: "Names of process environment variables the hook may read via its second `env` argument at evaluation time. Each name must match `^[A-Za-z_][A-Za-z0-9_]*$`. Variables that are unset on the connector or desktop process are omitted from `env`.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`), "environment variable name must match ^[A-Za-z_][A-Za-z0-9_]*$"),
				},
			},
			"allowlisted_network_hosts": {
				Description: "Hostnames and IP addresses the hook may contact at evaluation time. Schemes, paths, and ports are not accepted. All ports on each host are allowed.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
			"created_at": {
				Description: "When the hook was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the hook was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func getAllowlistedEnvironmentVariables(d *schema.ResourceData) ([]string, error) {
	raw, ok := d.Get("allowlisted_environment_variables").(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("error reading allowlisted_environment_variables")
	}
	return lo.Map(raw.List(), func(item any, _ int) string {
		return item.(string)
	}), nil
}

func getAllowlistedNetworkHosts(d *schema.ResourceData) ([]string, error) {
	raw, ok := d.Get("allowlisted_network_hosts").(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("error reading allowlisted_network_hosts")
	}
	return lo.Map(raw.List(), func(item any, _ int) string {
		return item.(string)
	}), nil
}

func resourceHookCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	allowlistedEnv, err := getAllowlistedEnvironmentVariables(d)
	if err != nil {
		return diag.FromErr(err)
	}
	allowlistedNetworkHosts, err := getAllowlistedNetworkHosts(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &corev1.CreateHookRequest{
		Name:                            d.Get("name").(string),
		Description:                     d.Get("description").(string),
		Code:                            d.Get("code").(string),
		Status:                          d.Get("status").(string),
		TimeoutMs:                       int32(d.Get("timeout_ms").(int)),
		AllowlistedEnvironmentVariables: allowlistedEnv,
		AllowlistedNetworkHosts:         allowlistedNetworkHosts,
	}

	res, err := c.Grpc.Sdk.HookServiceClient.CreateHook(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Hook.Id)

	return resourceHookRead(ctx, d, meta)
}

func resourceHookRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	hookID := d.Id()

	res, err := c.Grpc.Sdk.HookServiceClient.GetHook(ctx, connect.NewRequest(&corev1.GetHookRequest{Id: hookID}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Hook with ID "+hookID+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	hook := res.Msg.Hook
	d.Set("id", hook.Id)
	d.Set("name", hook.Name)
	d.Set("description", hook.Description)
	d.Set("code", hook.Code)
	d.Set("status", hook.Status)
	d.Set("timeout_ms", int(hook.TimeoutMs))
	d.Set("allowlisted_environment_variables", hook.AllowlistedEnvironmentVariables)
	d.Set("allowlisted_network_hosts", hook.AllowlistedNetworkHosts)
	if hook.CreatedAt != nil {
		d.Set("created_at", hook.CreatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	if hook.UpdatedAt != nil {
		d.Set("updated_at", hook.UpdatedAt.AsTime().UTC().Format(time.RFC3339))
	}

	return nil
}

func resourceHookUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("code") || d.HasChange("status") || d.HasChange("timeout_ms") || d.HasChange("allowlisted_environment_variables") || d.HasChange("allowlisted_network_hosts") {
		allowlistedEnv, err := getAllowlistedEnvironmentVariables(d)
		if err != nil {
			return diag.FromErr(err)
		}
		allowlistedNetworkHosts, err := getAllowlistedNetworkHosts(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req := &corev1.UpdateHookRequest{
			Hook: &corev1.Hook{
				Id:                              d.Id(),
				Name:                            d.Get("name").(string),
				Description:                     d.Get("description").(string),
				Code:                            d.Get("code").(string),
				Status:                          d.Get("status").(string),
				TimeoutMs:                       int32(d.Get("timeout_ms").(int)),
				AllowlistedEnvironmentVariables: allowlistedEnv,
				AllowlistedNetworkHosts:         allowlistedNetworkHosts,
			},
		}

		_, err = c.Grpc.Sdk.HookServiceClient.UpdateHook(ctx, connect.NewRequest(req))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceHookRead(ctx, d, meta)
}

func resourceHookDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.HookServiceClient.DeleteHook(ctx, connect.NewRequest(&corev1.DeleteHookRequest{Id: d.Id()}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
