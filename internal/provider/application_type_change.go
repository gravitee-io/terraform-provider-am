// Speakeasy lints its pristine output, before re-applying the region that calls this file;
// hack/scripts/check-gravitee-custom-regions.sh is what fails when the call goes missing.
//lint:file-ignore U1000 called from a hand-edit region of application_resource.go

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// omitUndeclaredDerivedOAuthSettings drops from an update the OAuth settings AM derives from the
// application type and kind, when one of those changes and the configuration leaves the setting unset.
//
// The generated Update sends the state for every value the plan leaves unknown, so a type change would
// re-send the settings derived from the previous type, and AM applies them over its own reset. Omitted,
// AM computes them for the new type, as it does for the Kubernetes operator.
//
// Called from the GRAVITEE DERIVED OAUTH SETTINGS region of ApplicationResource.Update.
func omitUndeclaredDerivedOAuthSettings(ctx context.Context, req resource.UpdateRequest, data *ApplicationResourceModel) diag.Diagnostics {
	changed, diags := typeOrKindChanged(ctx, req)
	if diags.HasError() || !changed || data.Settings == nil || data.Settings.Oauth == nil {
		return diags
	}

	oauth := data.Settings.Oauth
	declared := func(attribute string) bool {
		return isConfigured(req.Config.Raw, "settings", "oauth", attribute)
	}
	if !declared("force_pkce") {
		oauth.ForcePKCE = types.BoolNull()
	}
	if !declared("force_s256_code_challenge_method") {
		oauth.ForceS256CodeChallengeMethod = types.BoolNull()
	}
	if !declared("grant_types") {
		oauth.GrantTypes = nil
	}
	if !declared("response_types") {
		oauth.ResponseTypes = nil
	}
	if !declared("token_endpoint_auth_method") {
		oauth.TokenEndpointAuthMethod = types.StringNull()
	}
	return diags
}

func typeOrKindChanged(ctx context.Context, req resource.UpdateRequest) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	for _, attribute := range []string{"type", "kind"} {
		var planned, current types.String
		diags.Append(req.Plan.GetAttribute(ctx, path.Root(attribute), &planned)...)
		diags.Append(req.State.GetAttribute(ctx, path.Root(attribute), &current)...)
		if diags.HasError() {
			return false, diags
		}
		if !planned.Equal(current) {
			return true, diags
		}
	}
	return false, diags
}

// isConfigured reports whether the configuration sets the nested attribute named by attributes.
func isConfigured(config tftypes.Value, attributes ...string) bool {
	value := config
	for _, attribute := range attributes {
		if value.IsNull() {
			return false
		}
		if !value.IsKnown() {
			return true
		}
		var object map[string]tftypes.Value
		if err := value.As(&object); err != nil {
			return false
		}
		value = object[attribute]
	}
	return !value.IsNull()
}
