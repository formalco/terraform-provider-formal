package resource

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestOrgThemeValidationComesFromTheProto(t *testing.T) {
	logoTypes := orgThemeStringIn("logo_content_type")
	require.Len(t, logoTypes, 4)

	for _, tc := range []struct {
		field string
		value string
		valid bool
	}{
		{field: "logo_content_type", value: logoTypes[0], valid: true},
		{field: "logo_content_type", value: logoTypes[3], valid: true},
		{field: "logo_content_type", value: "image/gif"},
		{field: "logo_content_type", value: ""},
		{field: "display_name", value: ""},
		{field: "display_name", value: strings.Repeat("a", 60), valid: true},
		{field: "display_name", value: strings.Repeat("a", 61)},
	} {
		_, errs := orgThemeValidateString(tc.field)(tc.value, tc.field)
		require.Equal(t, tc.valid, len(errs) == 0, "%s %q: %v", tc.field, tc.value, errs)
	}
}

func TestValidateOrgThemeLogo(t *testing.T) {
	for _, tc := range []struct {
		logo  string
		valid bool
	}{
		{logo: "not base64!"},
		{logo: ""},
		{logo: base64.StdEncoding.EncodeToString(make([]byte, 256*1024+1))},
		{logo: base64.StdEncoding.EncodeToString([]byte("<svg/>")), valid: true},
	} {
		_, errs := validateOrgThemeLogo(tc.logo, "logo")
		require.Equal(t, tc.valid, len(errs) == 0, "%.20q: %v", tc.logo, errs)
	}
}

func TestOrgThemeRoundTrip(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceOrgTheme().Schema, map[string]any{
		"display_name":      "Acme",
		"logo":              base64.StdEncoding.EncodeToString([]byte("<svg/>")),
		"logo_content_type": "image/svg+xml",
	})
	args, err := expandOrgTheme(d)
	require.NoError(t, err)
	require.Equal(t, "Acme", args.displayName)
	require.Equal(t, "<svg/>", string(args.logo))
	require.Equal(t, "image/svg+xml", args.logoContentType)

	now := timestamppb.Now()
	logo := []byte(`<svg viewBox="0 0 2 2"/>`)
	require.NoError(t, flattenOrgTheme(d, &corev1.OrgTheme{
		Id:              "orgtheme_1",
		DisplayName:     "Acme Corp",
		Logo:            logo,
		LogoContentType: "image/svg+xml",
		CreatedAt:       now,
		UpdatedAt:       now,
	}))
	require.Equal(t, "orgtheme_1", d.Id())
	require.Equal(t, "Acme Corp", d.Get("display_name"))
	require.Equal(t, base64.StdEncoding.EncodeToString(logo), d.Get("logo"))
	require.Equal(t, "image/svg+xml", d.Get("logo_content_type"))
}
