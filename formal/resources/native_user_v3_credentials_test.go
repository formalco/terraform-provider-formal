package resource

import (
	"testing"

	"github.com/stretchr/testify/require"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func literalSecret(value string) []any {
	return []any{map[string]any{"literal": value, "environment_variable": ""}}
}

func environmentSecret(name string) []any {
	return []any{map[string]any{"literal": "", "environment_variable": name}}
}

func TestBuildNativeUserV3CredentialsCoversEveryVariant(t *testing.T) {
	tests := []struct {
		variant string
		block   map[string]any
		assert  func(t *testing.T, credentials *corev1.NativeUserV3Credentials)
	}{
		{
			variant: "basic",
			block:   map[string]any{"username": "app_admin", "password": literalSecret("s3cret")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				basic := credentials.GetBasic()
				require.Equal(t, "app_admin", basic.GetUsername())
				require.Equal(t, "s3cret", basic.GetPassword().GetLiteral())
			},
		},
		{
			variant: "aws_iam",
			block:   map[string]any{"username": "iam_user"},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				awsIam := credentials.GetAwsIam()
				require.Equal(t, "iam_user", awsIam.GetUsername())
			},
		},
		{
			variant: "aws_iam_role",
			block:   map[string]any{"username": "iam_user", "role": "arn:aws:iam::1:role/db"},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				awsIamRole := credentials.GetAwsIamRole()
				require.Equal(t, "iam_user", awsIamRole.GetUsername())
				require.Equal(t, "arn:aws:iam::1:role/db", awsIamRole.GetRole())
			},
		},
		{
			variant: "gcp_iam",
			block:   map[string]any{"username": "gcp_user"},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				require.Equal(t, "gcp_user", credentials.GetGcpIam().GetUsername())
			},
		},
		{
			variant: "azure_iam",
			block:   map[string]any{"username": "azure_user"},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				require.Equal(t, "azure_user", credentials.GetAzureIam().GetUsername())
			},
		},
		{
			variant: "kubernetes_path",
			block:   map[string]any{"kubeconfig_path": environmentSecret("KUBECONFIG")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				path := credentials.GetKubernetesPath().GetKubeconfigPath()
				require.Equal(t, "KUBECONFIG", path.GetEnvironmentVariable())
			},
		},
		{
			variant: "kubernetes_inline",
			block:   map[string]any{"kubeconfig": literalSecret("apiVersion: v1")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				require.Equal(t, "apiVersion: v1", credentials.GetKubernetesInline().GetKubeconfig().GetLiteral())
			},
		},
		{
			variant: "ssh_key",
			block: map[string]any{
				"username":    "ubuntu",
				"key":         literalSecret("PRIVATE KEY"),
				"certificate": literalSecret("SSH CERTIFICATE"),
			},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				sshKey := credentials.GetSshKey()
				require.Equal(t, "ubuntu", sshKey.GetUsername())
				require.Equal(t, "PRIVATE KEY", sshKey.GetKey().GetLiteral())
				require.Equal(t, "SSH CERTIFICATE", sshKey.GetCertificate().GetLiteral())
			},
		},
		{
			variant: "snowflake_key",
			block:   map[string]any{"username": "SVC", "key": environmentSecret("SNOWFLAKE_KEY")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				snowflake := credentials.GetSnowflakeKey()
				require.Equal(t, "SVC", snowflake.GetUsername())
				require.Equal(t, "SNOWFLAKE_KEY", snowflake.GetKey().GetEnvironmentVariable())
			},
		},
		{
			variant: "http_basic",
			block: map[string]any{
				"header": "Authorization", "username": "api", "password": literalSecret("pw"),
			},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				httpBasic := credentials.GetHttpBasic()
				require.Equal(t, "Authorization", httpBasic.GetHeader())
				require.Equal(t, "api", httpBasic.GetUsername())
				require.Equal(t, "pw", httpBasic.GetPassword().GetLiteral())
			},
		},
		{
			variant: "http_bearer",
			block:   map[string]any{"header": "Authorization", "token": literalSecret("tok")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				httpBearer := credentials.GetHttpBearer()
				require.Equal(t, "Authorization", httpBearer.GetHeader())
				require.Equal(t, "tok", httpBearer.GetToken().GetLiteral())
			},
		},
		{
			variant: "http_api_key_header",
			block:   map[string]any{"key": "X-Api-Key", "value": literalSecret("abc")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				apiKey := credentials.GetHttpApiKeyHeader()
				require.Equal(t, "X-Api-Key", apiKey.GetKey())
				require.Equal(t, "abc", apiKey.GetValue().GetLiteral())
			},
		},
		{
			variant: "http_api_key_query",
			block:   map[string]any{"key": "api_key", "value": environmentSecret("API_KEY")},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				apiKey := credentials.GetHttpApiKeyQuery()
				require.Equal(t, "api_key", apiKey.GetKey())
				require.Equal(t, "API_KEY", apiKey.GetValue().GetEnvironmentVariable())
			},
		},
		{
			variant: "hook",
			block: map[string]any{
				"code":                      "export default () => ({})",
				"output_type":               "basic",
				"allowlisted_env_variables": []any{"DB_USERNAME", "DB_PASSWORD"},
				"allowlisted_network_hosts": []any{"vault.internal"},
			},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				hook := credentials.GetHook()
				require.Equal(t, "export default () => ({})", hook.GetHook())
				require.Equal(t, "basic", hook.GetOutputType())
				require.Equal(t, []string{"DB_USERNAME", "DB_PASSWORD"}, hook.GetAllowlistedEnvVariables())
				require.Equal(t, []string{"vault.internal"}, hook.GetAllowlistedNetworkHosts())
			},
		},
	}

	require.Len(t, tests, len(nativeUserV3CredentialBlocks), "every credential variant must be covered")

	for _, test := range tests {
		t.Run(test.variant, func(t *testing.T) {
			credentials, err := buildNativeUserV3Credentials(test.variant, test.block)
			require.NoError(t, err)
			test.assert(t, credentials)
		})
	}
}

func TestBuildNativeUserV3CredentialsRejectsUnknownHookOutputType(t *testing.T) {
	_, err := buildNativeUserV3Credentials("hook", map[string]any{
		"code":        "export default () => ({})",
		"output_type": "not_a_type",
	})
	require.ErrorContains(t, err, "unsupported output_type")
}

func TestExpandSecretValueRequiresExactlyOneSource(t *testing.T) {
	_, err := expandSecretValue(map[string]any{
		"password": []any{map[string]any{"literal": "pw", "environment_variable": "PGPASSWORD"}},
	}, "password")
	require.ErrorContains(t, err, "set only one of literal or environment_variable")

	_, err = expandSecretValue(map[string]any{
		"password": []any{map[string]any{"literal": "", "environment_variable": ""}},
	}, "password")
	require.ErrorContains(t, err, "one of literal or environment_variable is required")

	_, err = expandSecretValue(map[string]any{}, "password")
	require.ErrorContains(t, err, "password is required")
}

func TestExpandOptionalSecretValueAllowsUnset(t *testing.T) {
	value, err := expandOptionalSecretValue(map[string]any{}, "certificate")
	require.NoError(t, err)
	require.Nil(t, value)
}

func TestDescribeNativeUserV3CredentialsKeepsRedactedLiteralFromState(t *testing.T) {
	priorLiteral := func(path string) string {
		require.Equal(t, "basic.0.password", path)
		return "the-secret-in-state"
	}

	variant, block, err := describeNativeUserV3Credentials(priorLiteral, &corev1.NativeUserV3Credentials{
		Value: &corev1.NativeUserV3Credentials_Basic{Basic: &corev1.BasicNativeUserV3{
			Username: "app_admin",
			Password: &corev1.SecretValue{Source: &corev1.SecretValue_Literal{
				Literal: nativeUserV3RedactedSecret,
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "basic", variant)
	require.Equal(t, "app_admin", block["username"])
	require.Equal(t, literalSecret("the-secret-in-state"), block["password"])
}

func TestDescribeNativeUserV3CredentialsReadsEnvironmentVariableSource(t *testing.T) {
	priorLiteral := func(string) string {
		t.Fatal("environment variable secrets must not read the literal from state")
		return ""
	}

	variant, block, err := describeNativeUserV3Credentials(priorLiteral, &corev1.NativeUserV3Credentials{
		Value: &corev1.NativeUserV3Credentials_HttpBearer{HttpBearer: &corev1.HTTPBearerNativeUserV3{
			Header: "Authorization",
			Token: &corev1.SecretValue{Source: &corev1.SecretValue_EnvironmentVariable{
				EnvironmentVariable: "API_TOKEN",
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "http_bearer", variant)
	require.Equal(t, environmentSecret("API_TOKEN"), block["token"])
}

func TestDescribeNativeUserV3CredentialsMapsHookOutputTypeBack(t *testing.T) {
	for _, name := range nativeUserV3TypeNames {
		variant, block, err := describeNativeUserV3Credentials(nil, &corev1.NativeUserV3Credentials{
			Value: &corev1.NativeUserV3Credentials_Hook{Hook: &corev1.HookNativeUserV3{
				Hook:                    "export default () => ({})",
				OutputType:              name,
				AllowlistedEnvVariables: []string{"DB_USERNAME"},
				AllowlistedNetworkHosts: []string{"vault.internal"},
			}},
		})
		require.NoError(t, err)
		require.Equal(t, "hook", variant)
		require.Equal(t, name, block["output_type"])
		require.Equal(t, []string{"DB_USERNAME"}, block["allowlisted_env_variables"])
		require.Equal(t, []string{"vault.internal"}, block["allowlisted_network_hosts"])
	}
}

func TestDescribeNativeUserV3CredentialsRejectsUnsetOneof(t *testing.T) {
	_, _, err := describeNativeUserV3Credentials(nil, &corev1.NativeUserV3Credentials{})
	require.ErrorContains(t, err, "does not support")
}
