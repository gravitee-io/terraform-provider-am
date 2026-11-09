// Sanity tests for the generated AM provider. These run without an APIM
// instance and only exercise the plugin-framework wiring that Speakeasy
// generated, plus the overlay-driven provider attributes. Their job is to
// catch regressions in overlays that quietly break resource registration
// or strip expected provider config.

package provider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gravitee-io/terraform-provider-am/internal/provider"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// newProvider returns the same provider instance the binary serves.
func newProvider() tfprovider.Provider {
	return provider.New("test")()
}

func TestProviderMetadata(t *testing.T) {
	t.Parallel()
	var resp tfprovider.MetadataResponse
	newProvider().Metadata(context.Background(), tfprovider.MetadataRequest{}, &resp)
	if resp.TypeName != "am" {
		t.Fatalf("expected TypeName \"am\", got %q", resp.TypeName)
	}
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()
	var resp tfprovider.SchemaResponse
	newProvider().Schema(context.Background(), tfprovider.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned diagnostics: %v", resp.Diagnostics)
	}

	wantOptional := []string{
		"server_url",
		"bearer_auth",
		"organization_id",
		"environment_id",
		"http_headers",
	}
	for _, name := range wantOptional {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("missing expected provider attribute %q", name)
			continue
		}
		if attr.IsRequired() {
			t.Errorf("provider attribute %q should be optional, not required", name)
		}
	}

	// Belt-and-braces: confirm bearer_auth is marked sensitive so we don't
	// regress on the gen.yaml security-feb-2025 fix.
	if bearer, ok := resp.Schema.Attributes["bearer_auth"]; ok {
		if str, isString := bearer.(schema.StringAttribute); isString && !str.Sensitive {
			t.Errorf("provider attribute bearer_auth should be Sensitive")
		}
	}
}

func TestProviderResourcesAndDataSources(t *testing.T) {
	t.Parallel()
	p := newProvider()
	ctx := context.Background()

	if got, want := len(p.Resources(ctx)), 7; got != want {
		t.Errorf("expected %d managed resources, got %d", want, got)
	}
	if got, want := len(p.DataSources(ctx)), 6; got != want {
		t.Errorf("expected %d data sources, got %d", want, got)
	}
}

// Every attribute that lands in a resource's URL path is its identity: changing it must replace
// the resource, not write a copy at the new path and leave the old one behind.
func TestResourceIdentityForcesReplacement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	identity := []string{"organization_id", "environment_id", "domain_key", "application_key", "key", "id"}

	for _, newResource := range newProvider().Resources(ctx) {
		r := newResource()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "am"}, &metadata)
		var resp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &resp)

		for _, name := range identity {
			attribute, ok := resp.Schema.Attributes[name].(resourceschema.StringAttribute)
			if !ok || attribute.Computed && !attribute.Optional {
				continue
			}
			if !forcesReplacement(ctx, attribute) {
				t.Errorf("%s.%s does not force replacement", metadata.TypeName, name)
			}
		}
	}
}

func forcesReplacement(ctx context.Context, attribute resourceschema.StringAttribute) bool {
	for _, modifier := range attribute.PlanModifiers {
		if strings.Contains(modifier.Description(ctx), "destroy and recreate") {
			return true
		}
	}
	return false
}
