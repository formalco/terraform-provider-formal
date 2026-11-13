package resource

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

// secretFixture is one secret as the schema defines it. Its zero value carries
// every source key, which is the shape flattenSecretValue emits and so what the
// tests that assert on it have to compare against.
type secretFixture struct {
	literal             string
	literalWO           string
	literalWOVersion    int
	environmentVariable string
	secretManager       map[string]any
}

func (s secretFixture) attrs() []any {
	secretManager := []any{}
	if s.secretManager != nil {
		secretManager = []any{s.secretManager}
	}
	return []any{map[string]any{
		"literal":              s.literal,
		"literal_wo":           s.literalWO,
		"literal_wo_version":   s.literalWOVersion,
		"environment_variable": s.environmentVariable,
		"secret_manager":       secretManager,
	}}
}

func secretManagerReference(awsARN, gcpName, azureURL string) map[string]any {
	provider := func(field, value string) []any {
		if value == "" {
			return []any{}
		}
		return []any{map[string]any{field: value}}
	}
	return map[string]any{
		"aws_secret":   provider("arn", awsARN),
		"gcp_secret":   provider("name", gcpName),
		"azure_secret": provider("url", azureURL),
	}
}

func noPriorSecret(string) map[string]any { return nil }

func literalSecret(value string) []any {
	return secretFixture{literal: value}.attrs()
}

func writeOnlySecret(value string, version int) []any {
	return secretFixture{literalWO: value, literalWOVersion: version}.attrs()
}

func environmentSecret(name string) []any {
	return secretFixture{environmentVariable: name}.attrs()
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
		{
			variant: "linked_http_basic",
			block:   map[string]any{"header": "Authorization", "username": "api"},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				linkedHttpBasic := credentials.GetLinkedHttpBasic()
				require.Equal(t, "Authorization", linkedHttpBasic.GetHeader())
				require.Equal(t, "api", linkedHttpBasic.GetUsername())
			},
		},
		{
			variant: "linked_http_bearer",
			block:   map[string]any{"header": "Authorization"},
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				require.Equal(t, "Authorization", credentials.GetLinkedHttpBearer().GetHeader())
			},
		},
		{
			variant: "passthrough",
			assert: func(t *testing.T, credentials *corev1.NativeUserV3Credentials) {
				require.NotNil(t, credentials.GetPassthrough())
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

func TestExpandNativeUserV3CredentialsReadsEmptyPassthroughBlock(t *testing.T) {
	d := schema.TestResourceDataRaw(t, nativeUserV3Schema(), map[string]any{
		"resource_id": "resource_1",
		"label":       "passthrough",
		"passthrough": []any{map[string]any{}},
	})

	credentials, err := expandNativeUserV3Credentials(d)
	require.NoError(t, err)
	require.NotNil(t, credentials.GetPassthrough())
}

func TestFlattenNativeUserV3CredentialsKeepsEmptyPassthroughBlock(t *testing.T) {
	d := schema.TestResourceDataRaw(t, nativeUserV3Schema(), map[string]any{})

	err := flattenNativeUserV3Credentials(d, &corev1.NativeUserV3Credentials{
		Value: &corev1.NativeUserV3Credentials_Passthrough{Passthrough: &corev1.PassthroughNativeUserV3{}},
	})
	require.NoError(t, err)
	require.Len(t, d.Get("passthrough"), 1)
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
		"password": secretFixture{literal: "pw", environmentVariable: "PGPASSWORD"}.attrs(),
	}, "password")
	require.ErrorContains(t, err, "set only one of literal, literal_wo, environment_variable or secret_manager")

	_, err = expandSecretValue(map[string]any{
		"password": secretFixture{literal: "pw", literalWO: "pw", literalWOVersion: 1}.attrs(),
	}, "password")
	require.ErrorContains(t, err, "set only one of literal, literal_wo, environment_variable or secret_manager")

	_, err = expandSecretValue(map[string]any{"password": secretFixture{}.attrs()}, "password")
	require.ErrorContains(t, err, "one of literal, literal_wo, environment_variable or secret_manager is required")

	_, err = expandSecretValue(map[string]any{}, "password")
	require.ErrorContains(t, err, "password is required")
}

func TestExpandSecretValueReadsSecretManagerReference(t *testing.T) {
	const (
		awsARN   = "arn:aws:secretsmanager:us-east-1:123456789012:secret:db-AbCdEf:password::"
		gcpName  = "projects/p/secrets/db/versions/latest"
		azureURL = "https://formal.vault.azure.net/secrets/db"
	)
	tests := []struct {
		name      string
		reference map[string]any
		want      *corev1.SecretManagerReference
	}{
		{
			name:      "aws",
			reference: secretManagerReference(awsARN, "", ""),
			want: &corev1.SecretManagerReference{Secret: &corev1.SecretManagerReference_AwsSecret{
				AwsSecret: &corev1.AWSSecretsManagerSecret{Arn: awsARN},
			}},
		},
		{
			name:      "gcp",
			reference: secretManagerReference("", gcpName, ""),
			want: &corev1.SecretManagerReference{Secret: &corev1.SecretManagerReference_GcpSecret{
				GcpSecret: &corev1.GCPSecretManagerSecret{Name: gcpName},
			}},
		},
		{
			name:      "azure",
			reference: secretManagerReference("", "", azureURL),
			want: &corev1.SecretManagerReference{Secret: &corev1.SecretManagerReference_AzureSecret{
				AzureSecret: &corev1.AzureKeyVaultSecret{Url: azureURL},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := expandSecretValue(map[string]any{
				"password": secretFixture{secretManager: test.reference}.attrs(),
			}, "password")
			require.NoError(t, err)
			require.True(t, proto.Equal(test.want, value.GetSecretManager()))

			_, block, err := describeNativeUserV3Credentials(noPriorSecret, &corev1.NativeUserV3Credentials{
				Value: &corev1.NativeUserV3Credentials_Basic{Basic: &corev1.BasicNativeUserV3{Password: value}},
			})
			require.NoError(t, err)
			require.Equal(t, secretFixture{secretManager: test.reference}.attrs(), block["password"])
		})
	}
}

func TestExpandSecretValueRequiresExactlyOneSecretManagerProvider(t *testing.T) {
	_, err := expandSecretValue(map[string]any{
		"password": secretFixture{secretManager: secretManagerReference("arn:aws:secretsmanager:x", "projects/p/secrets/s/versions/1", "")}.attrs(),
	}, "password")
	require.ErrorContains(t, err, "password.secret_manager: set exactly one of")

	_, err = expandSecretValue(map[string]any{
		"password": secretFixture{secretManager: secretManagerReference("", "", "")}.attrs(),
	}, "password")
	require.ErrorContains(t, err, "password.secret_manager: set exactly one of")

	_, err = expandSecretValue(map[string]any{
		"password": secretFixture{literal: "pw", secretManager: secretManagerReference("arn:aws:secretsmanager:x", "", "")}.attrs(),
	}, "password")
	require.ErrorContains(t, err, "set only one of literal, literal_wo, environment_variable or secret_manager")
}

func TestExpandNativeUserV3CredentialsReadsHookSecrets(t *testing.T) {
	d := schema.TestResourceDataRaw(t, nativeUserV3Schema(), map[string]any{
		"resource_id": "resource_1",
		"label":       "hook",
		"hook": []any{map[string]any{
			"code":        "export default (input, env, cache, secrets) => ({})",
			"output_type": "basic",
			"secret": []any{
				lo.Assign(secretManagerReference("arn:aws:secretsmanager:us-east-1:1:secret:db", "", ""), map[string]any{"name": "DB_PASSWORD"}),
				lo.Assign(secretManagerReference("", "", "https://formal.vault.azure.net/secrets/api"), map[string]any{"name": "API_KEY"}),
			},
		}},
	})

	credentials, err := expandNativeUserV3Credentials(d)
	require.NoError(t, err)
	secrets := credentials.GetHook().GetSecrets()
	require.Len(t, secrets, 2)
	require.Equal(t, "arn:aws:secretsmanager:us-east-1:1:secret:db", secrets["DB_PASSWORD"].GetAwsSecret().GetArn())
	require.Equal(t, "https://formal.vault.azure.net/secrets/api", secrets["API_KEY"].GetAzureSecret().GetUrl())

	require.NoError(t, flattenNativeUserV3Credentials(d, credentials))
	require.Equal(t, 2, d.Get("hook.0.secret").(*schema.Set).Len())
}

func TestFlattenHookSecretsDoesNotDriftFromConfig(t *testing.T) {
	resource := ResourceNativeUserV3()
	config := map[string]any{
		"resource_id": "resource_1",
		"label":       "hook",
		"hook": []any{map[string]any{
			"code":        "export default (input, env, cache, secrets) => ({})",
			"output_type": "basic",
			"secret": []any{
				map[string]any{"name": "A", "aws_secret": []any{map[string]any{"arn": "arn:aws:secretsmanager:us-east-1:1:secret:a"}}},
				map[string]any{"name": "B", "gcp_secret": []any{map[string]any{"name": "projects/p/secrets/b/versions/1"}}},
				map[string]any{"name": "C", "azure_secret": []any{map[string]any{"url": "https://formal.vault.azure.net/secrets/c"}}},
			},
		}},
	}

	d := schema.TestResourceDataRaw(t, nativeUserV3Schema(), config)
	credentials, err := expandNativeUserV3Credentials(d)
	require.NoError(t, err)

	applied := resource.Data(nil)
	applied.SetId("native_user_1")
	require.NoError(t, applied.Set("resource_id", "resource_1"))
	require.NoError(t, applied.Set("label", "hook"))
	require.NoError(t, flattenNativeUserV3Credentials(applied, credentials))

	diff, err := resource.SimpleDiff(t.Context(), applied.State(), terraform.NewResourceConfigRaw(config), nil)
	require.NoError(t, err)
	require.Empty(t, lo.PickBy(diff.Attributes, func(key string, _ *terraform.ResourceAttrDiff) bool {
		return strings.HasPrefix(key, "hook.")
	}))
}

func TestBuildNativeUserV3CredentialsRejectsDuplicateHookSecretNames(t *testing.T) {
	_, err := buildNativeUserV3Credentials("hook", map[string]any{
		"code":        "export default () => ({})",
		"output_type": "basic",
		"secret": []any{
			lo.Assign(secretManagerReference("arn:aws:secretsmanager:us-east-1:1:secret:a", "", ""), map[string]any{"name": "DB_PASSWORD"}),
			lo.Assign(secretManagerReference("arn:aws:secretsmanager:us-east-1:1:secret:b", "", ""), map[string]any{"name": "DB_PASSWORD"}),
		},
	})
	require.ErrorContains(t, err, `hook: secret "DB_PASSWORD" is set more than once`)
}

func TestFlattenHookSecretsSortsByName(t *testing.T) {
	flattened := flattenHookSecrets(map[string]*corev1.SecretManagerReference{
		"B": {Secret: &corev1.SecretManagerReference_GcpSecret{GcpSecret: &corev1.GCPSecretManagerSecret{Name: "projects/p/secrets/b/versions/1"}}},
		"A": {Secret: &corev1.SecretManagerReference_AwsSecret{AwsSecret: &corev1.AWSSecretsManagerSecret{Arn: "arn:aws:secretsmanager:us-east-1:1:secret:a"}}},
	})

	want := []any{
		lo.Assign(secretManagerReference("arn:aws:secretsmanager:us-east-1:1:secret:a", "", ""), map[string]any{"name": "A"}),
		lo.Assign(secretManagerReference("", "projects/p/secrets/b/versions/1", ""), map[string]any{"name": "B"}),
	}
	require.Equal(t, want, flattened)
}

func TestExpandSecretValueReadsWriteOnlyLiteral(t *testing.T) {
	value, err := expandSecretValue(map[string]any{
		"password": writeOnlySecret("s3cret", 1),
	}, "password")
	require.NoError(t, err)
	require.Equal(t, "s3cret", value.GetLiteral())
}

// secretAttrPath addresses one attribute of a variant's secret block.
func secretAttrPath(variant, field, attr string) cty.Path {
	return cty.GetAttrPath(variant).IndexInt(0).GetAttr(field).IndexInt(0).GetAttr(attr)
}

type rawConfigEntry struct {
	path  cty.Path
	value cty.Value
}

// fakeRawConfig answers only the paths it was given, failing the test on any
// other read so the tests pin exactly which paths are consulted.
func fakeRawConfig(t *testing.T, entries ...rawConfigEntry) rawConfigReader {
	t.Helper()
	return func(path cty.Path) (cty.Value, diag.Diagnostics) {
		for _, entry := range entries {
			if path.Equals(entry.path) {
				return entry.value, nil
			}
		}
		t.Fatalf("unexpected raw config path %#v", path)
		return cty.NilVal, nil
	}
}

// writeOnlySecretEntries covers both reads restoreWriteOnlySecrets makes for a
// secret: the write-only literal and its version trigger.
func writeOnlySecretEntries(variant, field, literal string, version int) []rawConfigEntry {
	return []rawConfigEntry{
		{secretAttrPath(variant, field, "literal_wo"), cty.StringVal(literal)},
		{secretAttrPath(variant, field, "literal_wo_version"), cty.NumberIntVal(int64(version))},
	}
}

func TestRestoreWriteOnlySecretsReadsEveryConfiguredSecretBlock(t *testing.T) {
	// Terraform hands the provider a blanked literal_wo, so the raw config is
	// the only place the value can come from.
	block := map[string]any{
		"username":    "ubuntu",
		"key":         writeOnlySecret("", 1),
		"certificate": writeOnlySecret("", 1),
	}

	entries := append(
		writeOnlySecretEntries("ssh_key", "key", "PRIVATE KEY", 1),
		writeOnlySecretEntries("ssh_key", "certificate", "SSH CERTIFICATE", 1)...,
	)
	restored, err := restoreWriteOnlySecrets(fakeRawConfig(t, entries...), "ssh_key", block)
	require.NoError(t, err)

	credentials, err := buildNativeUserV3Credentials("ssh_key", restored)
	require.NoError(t, err)
	require.Equal(t, "PRIVATE KEY", credentials.GetSshKey().GetKey().GetLiteral())
	require.Equal(t, "SSH CERTIFICATE", credentials.GetSshKey().GetCertificate().GetLiteral())

	// The SDK memoizes the map d.Get returns, so it must come back untouched.
	require.Empty(t, blockString(mustSingleBlock(t, block["key"]), "literal_wo"))
}

func TestRestoreWriteOnlySecretsRequiresVersionTrigger(t *testing.T) {
	block := map[string]any{"username": "app", "password": writeOnlySecret("", 0)}

	read := fakeRawConfig(t,
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo"), cty.StringVal("s3cret")},
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo_version"), cty.NullVal(cty.Number)},
	)

	_, err := restoreWriteOnlySecrets(read, "basic", block)
	require.ErrorContains(t, err, "basic.password: literal_wo requires literal_wo_version")
}

func TestRestoreWriteOnlySecretsAcceptsZeroVersion(t *testing.T) {
	// 0 is a configured version, matching the schema-level RequiredWith that
	// formal_native_user gets. Only an unset trigger is an error.
	block := map[string]any{"username": "app", "password": writeOnlySecret("", 0)}

	read := fakeRawConfig(t, writeOnlySecretEntries("basic", "password", "s3cret", 0)...)

	restored, err := restoreWriteOnlySecrets(read, "basic", block)
	require.NoError(t, err)
	credentials, err := buildNativeUserV3Credentials("basic", restored)
	require.NoError(t, err)
	require.Equal(t, "s3cret", credentials.GetBasic().GetPassword().GetLiteral())
}

func TestRestoreWriteOnlySecretsRejectsVersionTriggerWithoutLiteral(t *testing.T) {
	// Bumping the trigger with no literal_wo would resend the unchanged
	// credential and report a rotation that never happened.
	block := map[string]any{"username": "app", "password": environmentSecret("PGPASSWORD")}

	read := fakeRawConfig(t,
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo"), cty.NullVal(cty.String)},
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo_version"), cty.NumberIntVal(2)},
	)

	_, err := restoreWriteOnlySecrets(read, "basic", block)
	require.ErrorContains(t, err, "basic.password: literal_wo_version requires literal_wo")
}

func TestRestoreWriteOnlySecretsSkipsNonSecretFields(t *testing.T) {
	block := map[string]any{
		"code":                      "export default () => ({})",
		"output_type":               "basic",
		"allowlisted_env_variables": []any{"DB_PASSWORD"},
		"allowlisted_network_hosts": []any{"vault.internal"},
	}

	// fakeRawConfig with no entries fails on any read at all.
	_, err := restoreWriteOnlySecrets(fakeRawConfig(t), "hook", block)
	require.NoError(t, err)
}

func TestRestoreWriteOnlySecretsSkipsUnsetOptionalSecret(t *testing.T) {
	// An absent certificate block has no list element to address, so asking for
	// its path would make GetRawConfigAt fail.
	block := map[string]any{"username": "ubuntu", "key": writeOnlySecret("", 1)}

	read := fakeRawConfig(t, writeOnlySecretEntries("ssh_key", "key", "PRIVATE KEY", 1)...)

	_, err := restoreWriteOnlySecrets(read, "ssh_key", block)
	require.NoError(t, err)
}

func TestRestoreWriteOnlySecretsLeavesOtherSourcesAlone(t *testing.T) {
	block := map[string]any{"username": "app", "password": environmentSecret("PGPASSWORD")}

	read := fakeRawConfig(t,
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo"), cty.NullVal(cty.String)},
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo_version"), cty.NullVal(cty.Number)},
	)

	restored, err := restoreWriteOnlySecrets(read, "basic", block)
	require.NoError(t, err)

	credentials, err := buildNativeUserV3Credentials("basic", restored)
	require.NoError(t, err)
	require.Equal(t, "PGPASSWORD", credentials.GetBasic().GetPassword().GetEnvironmentVariable())
}

// An unknown value is neither null nor a type mismatch, so without the
// knownness check it would reach AsString and panic the plugin.
func TestRestoreWriteOnlySecretsRejectsUnknownLiteral(t *testing.T) {
	block := map[string]any{"username": "app", "password": writeOnlySecret("", 1)}

	read := fakeRawConfig(t,
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo"), cty.UnknownVal(cty.String)},
		rawConfigEntry{secretAttrPath("basic", "password", "literal_wo_version"), cty.NumberIntVal(1)},
	)

	_, err := restoreWriteOnlySecrets(read, "basic", block)
	require.ErrorContains(t, err, "basic.password: literal_wo must be a known string")
}

// A read failure has to abort rather than leave the secret blank.
func TestRestoreWriteOnlySecretsSurfacesRawConfigErrors(t *testing.T) {
	block := map[string]any{"username": "app", "password": writeOnlySecret("", 1)}

	read := func(cty.Path) (cty.Value, diag.Diagnostics) {
		return cty.NilVal, diag.Diagnostics{{Severity: diag.Error, Summary: "Invalid config path"}}
	}

	_, err := restoreWriteOnlySecrets(read, "basic", block)
	require.ErrorContains(t, err, "basic.password: failed to get literal_wo")
	require.Empty(t, blockString(mustSingleBlock(t, block["password"]), "literal_wo"))
}

func mustSingleBlock(t *testing.T, value any) map[string]any {
	t.Helper()
	block, ok := singleBlock(value)
	require.True(t, ok)
	return block
}

func TestExpandOptionalSecretValueAllowsUnset(t *testing.T) {
	value, err := expandOptionalSecretValue(map[string]any{}, "certificate")
	require.NoError(t, err)
	require.Nil(t, value)
}

func TestDescribeNativeUserV3CredentialsKeepsRedactedLiteralFromState(t *testing.T) {
	priorSecret := func(path string) map[string]any {
		require.Equal(t, "basic.0.password", path)
		return map[string]any{"literal": "the-secret-in-state"}
	}

	variant, block, err := describeNativeUserV3Credentials(priorSecret, &corev1.NativeUserV3Credentials{
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

func TestDescribeNativeUserV3CredentialsKeepsWriteOnlyVersionFromState(t *testing.T) {
	// A write-only literal is never in state, so the redacted value must not
	// land in `literal`, and the version trigger has to survive the read or
	// every plan would show drift.
	priorSecret := func(path string) map[string]any {
		require.Equal(t, "basic.0.password", path)
		return map[string]any{"literal": "", "literal_wo_version": 3}
	}

	_, block, err := describeNativeUserV3Credentials(priorSecret, &corev1.NativeUserV3Credentials{
		Value: &corev1.NativeUserV3Credentials_Basic{Basic: &corev1.BasicNativeUserV3{
			Username: "app_admin",
			Password: &corev1.SecretValue{Source: &corev1.SecretValue_Literal{
				Literal: nativeUserV3RedactedSecret,
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, writeOnlySecret("", 3), block["password"])
}

func TestDescribeNativeUserV3CredentialsReadsEnvironmentVariableSource(t *testing.T) {
	// State holds a literal left over from a previous source. An environment
	// variable secret must not pull it back in.
	priorSecret := func(path string) map[string]any {
		require.Equal(t, "http_bearer.0.token", path)
		return map[string]any{"literal": "leftover", "literal_wo_version": 7}
	}

	variant, block, err := describeNativeUserV3Credentials(priorSecret, &corev1.NativeUserV3Credentials{
		Value: &corev1.NativeUserV3Credentials_HttpBearer{HttpBearer: &corev1.HTTPBearerNativeUserV3{
			Header: "Authorization",
			Token: &corev1.SecretValue{Source: &corev1.SecretValue_EnvironmentVariable{
				EnvironmentVariable: "API_TOKEN",
			}},
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "http_bearer", variant)
	require.Equal(t, secretFixture{environmentVariable: "API_TOKEN", literalWOVersion: 7}.attrs(), block["token"])
}

func TestDescribeNativeUserV3CredentialsMapsHookOutputTypeBack(t *testing.T) {
	for _, name := range nativeUserV3TypeNames {
		variant, block, err := describeNativeUserV3Credentials(noPriorSecret, &corev1.NativeUserV3Credentials{
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
	_, _, err := describeNativeUserV3Credentials(noPriorSecret, &corev1.NativeUserV3Credentials{})
	require.ErrorContains(t, err, "does not support")
}
