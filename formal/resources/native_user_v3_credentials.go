package resource

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

// expandNativeUserV3Credentials builds the credentials oneof from whichever
// credential block the configuration set.
func expandNativeUserV3Credentials(d *schema.ResourceData) (*corev1.NativeUserV3Credentials, error) {
	for _, name := range nativeUserV3CredentialBlocks {
		block, ok := singleBlock(d.Get(name))
		if !ok {
			continue
		}
		return buildNativeUserV3Credentials(name, block)
	}
	return nil, fmt.Errorf(
		"exactly one credential block must be set, one of: %s",
		strings.Join(nativeUserV3CredentialBlocks, ", "),
	)
}

func buildNativeUserV3Credentials(variant string, block map[string]any) (*corev1.NativeUserV3Credentials, error) {
	switch variant {
	case "basic":
		password, err := expandSecretValue(block, "password")
		if err != nil {
			return nil, fmt.Errorf("basic: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_Basic{
			Basic: &corev1.BasicNativeUserV3{
				Username: blockString(block, "username"),
				Password: password,
			},
		}}, nil

	case "aws_iam":
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_AwsIam{
			AwsIam: &corev1.AWSIAMNativeUserV3{Username: blockString(block, "username")},
		}}, nil

	case "aws_iam_role":
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_AwsIamRole{
			AwsIamRole: &corev1.AWSIAMRoleNativeUserV3{
				Username: blockString(block, "username"),
				Role:     blockString(block, "role"),
			},
		}}, nil

	case "gcp_iam":
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_GcpIam{
			GcpIam: &corev1.GCPIAMNativeUserV3{Username: blockString(block, "username")},
		}}, nil

	case "azure_iam":
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_AzureIam{
			AzureIam: &corev1.AzureIAMNativeUserV3{Username: blockString(block, "username")},
		}}, nil

	case "kubernetes_path":
		path, err := expandSecretValue(block, "kubeconfig_path")
		if err != nil {
			return nil, fmt.Errorf("kubernetes_path: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_KubernetesPath{
			KubernetesPath: &corev1.KubernetesPathNativeUserV3{KubeconfigPath: path},
		}}, nil

	case "kubernetes_inline":
		kubeconfig, err := expandSecretValue(block, "kubeconfig")
		if err != nil {
			return nil, fmt.Errorf("kubernetes_inline: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_KubernetesInline{
			KubernetesInline: &corev1.KubernetesInlineNativeUserV3{Kubeconfig: kubeconfig},
		}}, nil

	case "ssh_key":
		key, err := expandSecretValue(block, "key")
		if err != nil {
			return nil, fmt.Errorf("ssh_key: %w", err)
		}
		certificate, err := expandOptionalSecretValue(block, "certificate")
		if err != nil {
			return nil, fmt.Errorf("ssh_key: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_SshKey{
			SshKey: &corev1.SSHKeyNativeUserV3{
				Username:    blockString(block, "username"),
				Key:         key,
				Certificate: certificate,
			},
		}}, nil

	case "snowflake_key":
		key, err := expandSecretValue(block, "key")
		if err != nil {
			return nil, fmt.Errorf("snowflake_key: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_SnowflakeKey{
			SnowflakeKey: &corev1.SnowflakeKeyNativeUserV3{
				Username: blockString(block, "username"),
				Key:      key,
			},
		}}, nil

	case "http_basic":
		password, err := expandSecretValue(block, "password")
		if err != nil {
			return nil, fmt.Errorf("http_basic: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_HttpBasic{
			HttpBasic: &corev1.HTTPBasicNativeUserV3{
				Header:   blockString(block, "header"),
				Username: blockString(block, "username"),
				Password: password,
			},
		}}, nil

	case "http_bearer":
		token, err := expandSecretValue(block, "token")
		if err != nil {
			return nil, fmt.Errorf("http_bearer: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_HttpBearer{
			HttpBearer: &corev1.HTTPBearerNativeUserV3{
				Header: blockString(block, "header"),
				Token:  token,
			},
		}}, nil

	case "http_api_key_header":
		value, err := expandSecretValue(block, "value")
		if err != nil {
			return nil, fmt.Errorf("http_api_key_header: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_HttpApiKeyHeader{
			HttpApiKeyHeader: &corev1.HTTPAPIKeyHeaderNativeUserV3{
				Key:   blockString(block, "key"),
				Value: value,
			},
		}}, nil

	case "http_api_key_query":
		value, err := expandSecretValue(block, "value")
		if err != nil {
			return nil, fmt.Errorf("http_api_key_query: %w", err)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_HttpApiKeyQuery{
			HttpApiKeyQuery: &corev1.HTTPAPIKeyQueryNativeUserV3{
				Key:   blockString(block, "key"),
				Value: value,
			},
		}}, nil

	case "hook":
		outputTypeName := blockString(block, "output_type")
		if !slices.Contains(nativeUserV3TypeNames, outputTypeName) {
			return nil, fmt.Errorf(
				"hook: unsupported output_type %q, supported values are %s",
				outputTypeName, strings.Join(nativeUserV3TypeNames, ", "),
			)
		}
		return &corev1.NativeUserV3Credentials{Value: &corev1.NativeUserV3Credentials_Hook{
			Hook: &corev1.HookNativeUserV3{
				Hook:                    blockString(block, "code"),
				OutputType:              outputTypeName,
				AllowlistedEnvVariables: blockStringList(block, "allowlisted_env_variables"),
				AllowlistedNetworkHosts: blockStringList(block, "allowlisted_network_hosts"),
			},
		}}, nil
	}

	return nil, fmt.Errorf("unsupported credential block %q", variant)
}

// flattenNativeUserV3Credentials writes the credentials the API returned into
// state, clearing the credential blocks that are not in use.
func flattenNativeUserV3Credentials(d *schema.ResourceData, credentials *corev1.NativeUserV3Credentials) error {
	priorLiteral := func(path string) string {
		literal, _ := d.Get(path + ".0.literal").(string)
		return literal
	}

	variant, block, err := describeNativeUserV3Credentials(priorLiteral, credentials)
	if err != nil {
		return err
	}

	for _, name := range nativeUserV3CredentialBlocks {
		if name != variant {
			if err := d.Set(name, []any{}); err != nil {
				return err
			}
			continue
		}
		if err := d.Set(name, []any{block}); err != nil {
			return err
		}
	}
	return nil
}

// priorLiteralLookup returns the literal secret already in state at a block path.
type priorLiteralLookup func(path string) string

func describeNativeUserV3Credentials(priorLiteral priorLiteralLookup, credentials *corev1.NativeUserV3Credentials) (string, map[string]any, error) {
	switch value := credentials.GetValue().(type) {
	case *corev1.NativeUserV3Credentials_Basic:
		return "basic", map[string]any{
			"username": value.Basic.GetUsername(),
			"password": flattenSecretValue(priorLiteral, "basic.0.password", value.Basic.GetPassword()),
		}, nil

	case *corev1.NativeUserV3Credentials_AwsIam:
		return "aws_iam", map[string]any{
			"username": value.AwsIam.GetUsername(),
		}, nil

	case *corev1.NativeUserV3Credentials_AwsIamRole:
		return "aws_iam_role", map[string]any{
			"username": value.AwsIamRole.GetUsername(),
			"role":     value.AwsIamRole.GetRole(),
		}, nil

	case *corev1.NativeUserV3Credentials_GcpIam:
		return "gcp_iam", map[string]any{
			"username": value.GcpIam.GetUsername(),
		}, nil

	case *corev1.NativeUserV3Credentials_AzureIam:
		return "azure_iam", map[string]any{
			"username": value.AzureIam.GetUsername(),
		}, nil

	case *corev1.NativeUserV3Credentials_KubernetesPath:
		return "kubernetes_path", map[string]any{
			"kubeconfig_path": flattenSecretValue(priorLiteral, "kubernetes_path.0.kubeconfig_path", value.KubernetesPath.GetKubeconfigPath()),
		}, nil

	case *corev1.NativeUserV3Credentials_KubernetesInline:
		return "kubernetes_inline", map[string]any{
			"kubeconfig": flattenSecretValue(priorLiteral, "kubernetes_inline.0.kubeconfig", value.KubernetesInline.GetKubeconfig()),
		}, nil

	case *corev1.NativeUserV3Credentials_SshKey:
		block := map[string]any{
			"username": value.SshKey.GetUsername(),
			"key":      flattenSecretValue(priorLiteral, "ssh_key.0.key", value.SshKey.GetKey()),
		}
		if value.SshKey.GetCertificate() != nil {
			block["certificate"] = flattenSecretValue(priorLiteral, "ssh_key.0.certificate", value.SshKey.GetCertificate())
		}
		return "ssh_key", block, nil

	case *corev1.NativeUserV3Credentials_SnowflakeKey:
		return "snowflake_key", map[string]any{
			"username": value.SnowflakeKey.GetUsername(),
			"key":      flattenSecretValue(priorLiteral, "snowflake_key.0.key", value.SnowflakeKey.GetKey()),
		}, nil

	case *corev1.NativeUserV3Credentials_HttpBasic:
		return "http_basic", map[string]any{
			"header":   value.HttpBasic.GetHeader(),
			"username": value.HttpBasic.GetUsername(),
			"password": flattenSecretValue(priorLiteral, "http_basic.0.password", value.HttpBasic.GetPassword()),
		}, nil

	case *corev1.NativeUserV3Credentials_HttpBearer:
		return "http_bearer", map[string]any{
			"header": value.HttpBearer.GetHeader(),
			"token":  flattenSecretValue(priorLiteral, "http_bearer.0.token", value.HttpBearer.GetToken()),
		}, nil

	case *corev1.NativeUserV3Credentials_HttpApiKeyHeader:
		return "http_api_key_header", map[string]any{
			"key":   value.HttpApiKeyHeader.GetKey(),
			"value": flattenSecretValue(priorLiteral, "http_api_key_header.0.value", value.HttpApiKeyHeader.GetValue()),
		}, nil

	case *corev1.NativeUserV3Credentials_HttpApiKeyQuery:
		return "http_api_key_query", map[string]any{
			"key":   value.HttpApiKeyQuery.GetKey(),
			"value": flattenSecretValue(priorLiteral, "http_api_key_query.0.value", value.HttpApiKeyQuery.GetValue()),
		}, nil

	case *corev1.NativeUserV3Credentials_Hook:
		return "hook", map[string]any{
			"code":                      value.Hook.GetHook(),
			"output_type":               value.Hook.GetOutputType(),
			"allowlisted_env_variables": value.Hook.GetAllowlistedEnvVariables(),
			"allowlisted_network_hosts": value.Hook.GetAllowlistedNetworkHosts(),
		}, nil
	}

	return "", nil, fmt.Errorf("the API returned credentials this provider version does not support; please upgrade the Formal provider")
}

// flattenSecretValue writes a secret back into state. The API redacts literal
// secrets, so the value already in state is kept rather than overwritten with the
// redaction placeholder, which would otherwise show up as permanent drift.
func flattenSecretValue(priorLiteral priorLiteralLookup, path string, value *corev1.SecretValue) []any {
	secret := map[string]any{"literal": "", "environment_variable": ""}

	switch source := value.GetSource().(type) {
	case *corev1.SecretValue_EnvironmentVariable:
		secret["environment_variable"] = source.EnvironmentVariable
	case *corev1.SecretValue_Literal:
		if source.Literal == nativeUserV3RedactedSecret {
			secret["literal"] = priorLiteral(path)
		} else {
			secret["literal"] = source.Literal
		}
	}

	return []any{secret}
}

func expandSecretValue(block map[string]any, field string) (*corev1.SecretValue, error) {
	secret, ok := singleBlock(block[field])
	if !ok {
		return nil, fmt.Errorf("%s is required", field)
	}

	literal := blockString(secret, "literal")
	environmentVariable := blockString(secret, "environment_variable")

	switch {
	case literal != "" && environmentVariable != "":
		return nil, fmt.Errorf("%s: set only one of literal or environment_variable", field)
	case literal != "":
		return &corev1.SecretValue{Source: &corev1.SecretValue_Literal{Literal: literal}}, nil
	case environmentVariable != "":
		return &corev1.SecretValue{Source: &corev1.SecretValue_EnvironmentVariable{
			EnvironmentVariable: environmentVariable,
		}}, nil
	}

	return nil, fmt.Errorf("%s: one of literal or environment_variable is required", field)
}

func expandOptionalSecretValue(block map[string]any, field string) (*corev1.SecretValue, error) {
	if _, ok := singleBlock(block[field]); !ok {
		return nil, nil
	}
	return expandSecretValue(block, field)
}

// singleBlock reads a MaxItems-1 block into its attribute map.
func singleBlock(value any) (map[string]any, bool) {
	list, ok := value.([]any)
	if !ok || len(list) == 0 || list[0] == nil {
		return nil, false
	}
	block, ok := list[0].(map[string]any)
	return block, ok
}

func blockString(block map[string]any, field string) string {
	value, _ := block[field].(string)
	return value
}

func blockStringList(block map[string]any, field string) []string {
	values, _ := block[field].([]any)
	return lo.FilterMap(values, func(value any, _ int) (string, bool) {
		text, ok := value.(string)
		return text, ok
	})
}
