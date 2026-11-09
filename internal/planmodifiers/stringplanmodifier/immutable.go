package stringplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Immutable returns a plan modifier that rejects a change to the attribute on an existing resource.
// Any value is accepted on creation. An unknown plan value (attribute dropped from the configuration)
// is left as is: pair Immutable with a modifier that keeps the state value, such as SuppressDiff.
func Immutable() planmodifier.String {
	return immutable{}
}

type immutable struct{}

func (m immutable) Description(_ context.Context) string {
	return "Value cannot be changed once set."
}

func (m immutable) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutable) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Immutable attribute",
		req.Path.String()+" cannot be changed once set",
	)
}
