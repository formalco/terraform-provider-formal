package resource

import (
	"context"
	"strings"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSshHostKey() *schema.Resource {
	return &schema.Resource{
		Description: "Pin an upstream SSH host public key for a Formal resource. Multiple pins may exist for one resource_id when the host presents multiple key types.",

		CreateContext: resourceSshHostKeyCreate,
		ReadContext:   resourceSshHostKeyRead,
		UpdateContext: resourceSshHostKeyUpdate,
		DeleteContext: resourceSshHostKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the SSH host key pin.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "Resource ID for which the SSH host key is pinned.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"public_key": {
				Description: "OpenSSH public key of the upstream SSH host (for example the output of `ssh-keygen -yf /etc/ssh/ssh_host_ed25519_key`).",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceSshHostKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	msg := &corev1.CreateResourceSshHostKeyRequest{
		ResourceId: d.Get("resource_id").(string),
		PublicKey:  d.Get("public_key").(string),
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceSshHostKey(ctx, msg)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ResourceSshHostKey.Id)

	return resourceSshHostKeyRead(ctx, d, meta)
}

func resourceSshHostKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceSshHostKey(ctx, &corev1.GetResourceSshHostKeyRequest{Id: d.Id()})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The SSH host key was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("resource_id", res.ResourceSshHostKey.ResourceId)
	d.Set("public_key", res.ResourceSshHostKey.PublicKey)
	d.SetId(res.ResourceSshHostKey.Id)

	return diags
}

func resourceSshHostKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	fieldsThatCanChange := []string{"public_key"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceSshHostKey(ctx, &corev1.UpdateResourceSshHostKeyRequest{
		Id:        d.Id(),
		PublicKey: d.Get("public_key").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSshHostKeyRead(ctx, d, meta)
}

func resourceSshHostKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceSshHostKey(ctx, &corev1.DeleteResourceSshHostKeyRequest{Id: d.Id()})
	if err != nil {
		tflog.Warn(ctx, err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
