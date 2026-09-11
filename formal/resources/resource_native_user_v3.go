package resource

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

// nativeUserV3RedactedSecret is the placeholder the API returns in place of a
// literal secret. Sending it back on update keeps the stored secret unchanged.
const nativeUserV3RedactedSecret = "NOT_RETURNED"

// nativeUserV3CredentialBlocks lists every credential variant block, in proto
// field order. Exactly one must be set, and which one is set is immutable.
var nativeUserV3CredentialBlocks = []string{
	"basic",
	"aws_iam",
	"aws_iam_role",
	"gcp_iam",
	"azure_iam",
	"kubernetes_path",
	"kubernetes_inline",
	"ssh_key",
	"snowflake_key",
	"http_basic",
	"http_bearer",
	"http_api_key_header",
	"http_api_key_query",
	"hook",
}

var nativeUserV3TypeNames = []string{
	"basic",
	"aws_iam",
	"aws_iam_role",
	"gcp_iam",
	"azure_iam",
	"kubernetes_path",
	"kubernetes_inline",
	"ssh_key",
	"snowflake_key",
	"http_basic",
	"http_bearer",
	"http_api_key_header",
	"http_api_key_query",
}

func ResourceNativeUserV3() *schema.Resource {
	return &schema.Resource{
		Description: "A Native User the Formal connector authenticates to a Resource as. The Resource decides which Native User a session uses through its `formal_resource_native_user_selection` expression.",

		CreateContext: resourceNativeUserV3Create,
		ReadContext:   resourceNativeUserV3Read,
		UpdateContext: resourceNativeUserV3Update,
		DeleteContext: resourceNativeUserV3Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		// The API rejects changing which credential variant a Native User uses, so a
		// variant swap has to be a replacement rather than an update.
		CustomizeDiff: resourceNativeUserV3ForceNewOnVariantChange,

		Schema: nativeUserV3Schema(),
	}
}

func nativeUserV3Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Description: "The ID of the Native User.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"resource_id": {
			Description: "The ID of the Resource this Native User authenticates to.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"label": {
			Description: "A name for this Native User, unique within the Resource. Selection expressions reference the Native User by ID, so the label is safe to rename.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"termination_protection": {
			Description: "If set to true, this Native User cannot be deleted.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"created_at": {
			Description: "Creation time of the Native User.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"updated_at": {
			Description: "Last update time of the Native User.",
			Type:        schema.TypeString,
			Computed:    true,
		},

		"basic": credentialBlock(
			"Username and password authentication, for example Postgres or MySQL.",
			map[string]*schema.Schema{
				"username": requiredString("The username to authenticate as."),
				"password": secretBlock("The password to authenticate with."),
			},
		),
		"aws_iam": credentialBlock(
			"AWS IAM authentication.",
			map[string]*schema.Schema{
				"username": requiredString("The database username mapped to the IAM identity."),
			},
		),
		"aws_iam_role": credentialBlock(
			"AWS IAM authentication using an assumed role.",
			map[string]*schema.Schema{
				"username": requiredString("The database username mapped to the IAM identity."),
				"role":     requiredString("The ARN of the role to assume."),
			},
		),
		"gcp_iam": credentialBlock(
			"GCP IAM authentication.",
			map[string]*schema.Schema{
				"username": requiredString("The database username mapped to the GCP identity."),
			},
		),
		"azure_iam": credentialBlock(
			"Azure IAM authentication.",
			map[string]*schema.Schema{
				"username": requiredString("The database username mapped to the Azure identity."),
			},
		),
		"kubernetes_path": credentialBlock(
			"Kubernetes authentication using a kubeconfig file path on the connector.",
			map[string]*schema.Schema{
				"kubeconfig_path": secretBlock("Path to the kubeconfig file on the connector."),
			},
		),
		"kubernetes_inline": credentialBlock(
			"Kubernetes authentication using an inline kubeconfig document.",
			map[string]*schema.Schema{
				"kubeconfig": secretBlock("The kubeconfig YAML document."),
			},
		),
		"ssh_key": credentialBlock(
			"SSH private key authentication.",
			map[string]*schema.Schema{
				"username":    requiredString("The username to authenticate as."),
				"key":         secretBlock("The SSH private key."),
				"certificate": optionalSecretBlock("The optional SSH certificate paired with the private key."),
			},
		),
		"snowflake_key": credentialBlock(
			"Snowflake key-pair authentication.",
			map[string]*schema.Schema{
				"username": requiredString("The Snowflake username to authenticate as."),
				"key":      secretBlock("The private key."),
			},
		),
		"http_basic": credentialBlock(
			"HTTP Basic authentication, injected on a named header.",
			map[string]*schema.Schema{
				"header":   requiredString("The header to inject the credentials on, for example `Authorization`."),
				"username": requiredString("The username to authenticate as."),
				"password": secretBlock("The password to authenticate with."),
			},
		),
		"http_bearer": credentialBlock(
			"HTTP Bearer token authentication, injected on a named header.",
			map[string]*schema.Schema{
				"header": requiredString("The header to inject the token on, for example `Authorization`."),
				"token":  secretBlock("The bearer token."),
			},
		),
		"http_api_key_header": credentialBlock(
			"HTTP API key authentication, sent as a header.",
			map[string]*schema.Schema{
				"key":   requiredString("The name of the header carrying the API key."),
				"value": secretBlock("The API key."),
			},
		),
		"http_api_key_query": credentialBlock(
			"HTTP API key authentication, sent as a query parameter.",
			map[string]*schema.Schema{
				"key":   requiredString("The name of the query parameter carrying the API key."),
				"value": secretBlock("The API key."),
			},
		),
		"hook": credentialBlock(
			"Credentials resolved on the connector by running a hook.",
			map[string]*schema.Schema{
				"code": {
					Description: "The TypeScript or JavaScript source of the hook, in the same form as `formal_hook.code`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"output_type": {
					Description: fmt.Sprintf(
						"The credential shape the hook must return, so the connector can validate its result. One of: `%s`.",
						strings.Join(nativeUserV3TypeNames, "`, `"),
					),
					Type:     schema.TypeString,
					Required: true,
					// Part of the Native User's credential type, which the API treats as immutable.
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(nativeUserV3TypeNames, false),
				},
				"allowlisted_env_variables": {
					Description: "Environment variables the hook may read.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`),
							"environment variable name must match ^[A-Za-z_][A-Za-z0-9_]*$",
						),
					},
				},
				"allowlisted_network_hosts": {
					Description: "Network hosts the hook may access.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringIsNotEmpty,
					},
				},
			},
		),
	}

	return s
}

// credentialBlock wraps a credential variant's fields in a single-item block that
// is mutually exclusive with every other variant.
func credentialBlock(description string, fields map[string]*schema.Schema) *schema.Schema {
	return &schema.Schema{
		Description:  description,
		Type:         schema.TypeList,
		Optional:     true,
		MaxItems:     1,
		ExactlyOneOf: nativeUserV3CredentialBlocks,
		Elem:         &schema.Resource{Schema: fields},
	}
}

func requiredString(description string) *schema.Schema {
	return &schema.Schema{
		Description: description,
		Type:        schema.TypeString,
		Required:    true,
	}
}

// secretBlock describes a secret sourced either from a literal value or from an
// environment variable read on the connector. Exactly one source must be set;
// that is enforced when the block is expanded because Terraform cannot express
// mutual exclusion inside a nested block.
func secretBlock(description string) *schema.Schema {
	return secretBlockSchema(description, true)
}

func optionalSecretBlock(description string) *schema.Schema {
	return secretBlockSchema(description, false)
}

func secretBlockSchema(description string, required bool) *schema.Schema {
	return &schema.Schema{
		Description: description + " Set exactly one of `literal` or `environment_variable`.",
		Type:        schema.TypeList,
		Required:    required,
		Optional:    !required,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"literal": {
					Description: "The secret value itself. Stored in Terraform state; prefer `environment_variable` where possible.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
				},
				"environment_variable": {
					Description: "The name of an environment variable the connector reads the secret from.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
	}
}

// resourceNativeUserV3ForceNewOnVariantChange replaces the Native User when the
// configuration moves to a different credential variant.
func resourceNativeUserV3ForceNewOnVariantChange(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if d.Id() == "" {
		return nil
	}

	var before, after string
	for _, name := range nativeUserV3CredentialBlocks {
		oldValue, newValue := d.GetChange(name)
		if blockIsSet(oldValue) {
			before = name
		}
		if blockIsSet(newValue) {
			after = name
		}
	}

	if before != "" && after != "" && before != after {
		return d.ForceNew(after)
	}
	return nil
}

func blockIsSet(value any) bool {
	list, ok := value.([]any)
	return ok && len(list) > 0 && list[0] != nil
}

func resourceNativeUserV3Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	credentials, err := expandNativeUserV3Credentials(d)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.NativeUserServiceClient.CreateNativeUserV3(ctx, &corev1.CreateNativeUserV3Request{
		ResourceId:            d.Get("resource_id").(string),
		Label:                 d.Get("label").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
		Credentials:           credentials,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.NativeUser.Id)

	return resourceNativeUserV3Read(ctx, d, meta)
}

func resourceNativeUserV3Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	id := d.Id()

	res, err := c.Grpc.Sdk.NativeUserServiceClient.GetNativeUserV3(ctx, &corev1.GetNativeUserV3Request{Id: id})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Native User "+id+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	nativeUser := res.NativeUser

	d.Set("resource_id", nativeUser.ResourceId)
	d.Set("label", nativeUser.Label)
	d.Set("termination_protection", nativeUser.TerminationProtection)
	if nativeUser.CreatedAt != nil {
		d.Set("created_at", nativeUser.CreatedAt.AsTime().Format(time.RFC3339))
	}
	if nativeUser.UpdatedAt != nil {
		d.Set("updated_at", nativeUser.UpdatedAt.AsTime().Format(time.RFC3339))
	}

	if err := flattenNativeUserV3Credentials(d, nativeUser.Credentials); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(nativeUser.Id)

	return diags
}

func resourceNativeUserV3Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	req := &corev1.UpdateNativeUserV3Request{Id: d.Id()}

	if d.HasChange("label") {
		label := d.Get("label").(string)
		req.Label = &label
	}

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		req.TerminationProtection = &terminationProtection
	}

	// Credentials are replaced wholesale, so they are only sent when something in
	// the active variant changed. Literal secrets the API redacted read back as
	// NOT_RETURNED, which it accepts as "keep the stored value".
	if d.HasChanges(nativeUserV3CredentialBlocks...) {
		credentials, err := expandNativeUserV3Credentials(d)
		if err != nil {
			return diag.FromErr(err)
		}
		req.Credentials = credentials
	}

	if _, err := c.Grpc.Sdk.NativeUserServiceClient.UpdateNativeUserV3(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	return resourceNativeUserV3Read(ctx, d, meta)
}

func resourceNativeUserV3Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("Native User cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.NativeUserServiceClient.DeleteNativeUserV3(ctx, &corev1.DeleteNativeUserV3Request{Id: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
