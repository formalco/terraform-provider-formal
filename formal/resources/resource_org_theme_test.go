package resource

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestOrgThemeValidationComesFromTheProto(t *testing.T) {
	logoTypes := orgThemeStringIn("logo_content_type")
	if len(logoTypes) != 4 {
		t.Fatalf("expected the four logo types, got %v", logoTypes)
	}
	validateType := orgThemeValidateString("logo_content_type")
	for _, logoType := range logoTypes {
		if _, errs := validateType(logoType, "logo_content_type"); len(errs) > 0 {
			t.Fatalf("logo type %q from the proto was rejected: %v", logoType, errs)
		}
	}
	if _, errs := validateType("image/gif", "logo_content_type"); len(errs) == 0 {
		t.Fatal("expected image/gif to be rejected")
	}

	if _, errs := validateType("", "logo_content_type"); len(errs) == 0 {
		t.Fatal("expected an empty logo type to be rejected")
	}

	validateName := orgThemeValidateString("display_name")
	if _, errs := validateName("", "display_name"); len(errs) == 0 {
		t.Fatal("expected an empty display name to be rejected")
	}
	if _, errs := validateName(strings.Repeat("a", 60), "display_name"); len(errs) > 0 {
		t.Fatalf("a 60-character display name was rejected: %v", errs)
	}
	if _, errs := validateName(strings.Repeat("a", 61), "display_name"); len(errs) == 0 {
		t.Fatal("expected a 61-character display name to be rejected")
	}
}

func TestValidateOrgThemeLogo(t *testing.T) {
	if _, errs := validateOrgThemeLogo("not base64!", "logo"); len(errs) == 0 {
		t.Fatal("expected invalid base64 to be rejected")
	}
	if _, errs := validateOrgThemeLogo("", "logo"); len(errs) == 0 {
		t.Fatal("expected an empty logo to be rejected")
	}
	tooLarge := base64.StdEncoding.EncodeToString(make([]byte, 262145))
	if _, errs := validateOrgThemeLogo(tooLarge, "logo"); len(errs) == 0 {
		t.Fatal("expected a logo over 256 KiB to be rejected")
	}
	ok := base64.StdEncoding.EncodeToString([]byte("<svg/>"))
	if _, errs := validateOrgThemeLogo(ok, "logo"); len(errs) > 0 {
		t.Fatalf("expected a small logo to pass: %v", errs)
	}
}

func TestOrgThemeRoundTrip(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceOrgTheme().Schema, map[string]any{
		"display_name":      "Acme",
		"logo":              base64.StdEncoding.EncodeToString([]byte("<svg/>")),
		"logo_content_type": "image/svg+xml",
	})
	args, err := expandOrgTheme(d)
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}
	if args.displayName != "Acme" || string(args.logo) != "<svg/>" || args.logoContentType != "image/svg+xml" {
		t.Fatalf("unexpected expanded args: %+v", args)
	}

	now := timestamppb.Now()
	err = flattenOrgTheme(d, &corev1.OrgTheme{
		Id:              "orgtheme_1",
		DisplayName:     "Acme Corp",
		Logo:            []byte("<svg viewBox=\"0 0 2 2\"/>"),
		LogoContentType: "image/svg+xml",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		t.Fatalf("unexpected flatten error: %v", err)
	}
	if d.Id() != "orgtheme_1" || d.Get("display_name") != "Acme Corp" {
		t.Fatalf("unexpected flattened state: id=%s display_name=%v", d.Id(), d.Get("display_name"))
	}
	if d.Get("logo") != base64.StdEncoding.EncodeToString([]byte("<svg viewBox=\"0 0 2 2\"/>")) || d.Get("logo_content_type") != "image/svg+xml" {
		t.Fatalf("unexpected flattened logo: %v %v", d.Get("logo"), d.Get("logo_content_type"))
	}
}
