package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNativeRoleLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "This resource creates assigns a Native Role to a Formal Identity.",

		CreateContext: resourceNativeRoleLinkCreate,
		ReadContext:   resourceNativeRoleLinkRead,
		// UpdateContext: resourceNativeRoleLinkUpdate,
		DeleteContext: resourceNativeRoleLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Sidecar ID of the Native Role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"native_role_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Native Role ID of the Native Role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"formal_identity_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for the Role or Group to be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"formal_identity_type": {
				// This description is used by the documentation generator and the language server.
				Description: "The type of Formal Identity to be linked. Accepted values are `role` and `group`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

const terraformIdDelimiter = "#_#"

func resourceNativeRoleLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	datastoreId := d.Get("datastore_id").(string)
	nativeRoleId := d.Get("native_role_id").(string)
	formalIdentityId := d.Get("formal_identity_id").(string)
	formalIdentityType := d.Get("formal_identity_type").(string)

	_, err := c.Grpc.Sdk.NativeUserServiceClient.CreateNativeUserIdentityLink(ctx, connect.NewRequest(&adminv1.CreateNativeUserIdentityLinkRequest{
		DataStoreId:        datastoreId,
		NativeUserId:       nativeRoleId,
		IdentityId:         formalIdentityId,
		FormalIdentityType: formalIdentityType,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	terraformResourceId := datastoreId + terraformIdDelimiter + formalIdentityId
	d.SetId(terraformResourceId)

	resourceNativeRoleLinkRead(ctx, d, meta)
	return diags
}

func resourceNativeRoleLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	tfId := d.Id()
	// Split
	tfIdSplit := strings.Split(tfId, roleLinkGroupTerraformIdDelimiter)
	if len(tfIdSplit) != 2 {
		return diag.FromErr(errors.New("the Terraform Resource ID for Native Role Link is malformatted. Please contact Formal support"))
	}
	datastoreId := tfIdSplit[0]
	formalIdentityId := tfIdSplit[1]

	res, err := c.Grpc.Sdk.NativeUserServiceClient.GetNativeUserIdentityLink(ctx, connect.NewRequest(&adminv1.GetNativeUserIdentityLinkRequest{
		DataStoreId: datastoreId,
		IdentityId:  formalIdentityId,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Link was deleted
			tflog.Warn(ctx, "The Native Role Link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of
	d.Set("datastore_id", datastoreId)
	d.Set("native_role_id", res.Msg.Link.NativeUserId)
	d.Set("formal_identity_id", formalIdentityId)
	d.Set("formal_identity_type", res.Msg.Link.FormalIdentityType)

	d.SetId(tfId)

	return diags
}

func resourceNativeRoleLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	roleLinkGroupTerraformId := d.Id()
	// Split
	roleLinkGroupTerraformIdSplit := strings.Split(roleLinkGroupTerraformId, roleLinkGroupTerraformIdDelimiter)
	if len(roleLinkGroupTerraformIdSplit) != 2 {
		return diag.FromErr(errors.New("formal Terraform resource id for role_link_group is malformatted. Please contact Formal support"))
	}
	datastoreId := roleLinkGroupTerraformIdSplit[0]
	formalIdentityId := roleLinkGroupTerraformIdSplit[1]

	_, err := c.Grpc.Sdk.NativeUserServiceClient.DeleteNativeUserIdentityLink(ctx, connect.NewRequest(&adminv1.DeleteNativeUserIdentityLinkRequest{DataStoreId: datastoreId, IdentityId: formalIdentityId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
