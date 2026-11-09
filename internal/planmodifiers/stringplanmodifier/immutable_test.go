package stringplanmodifier_test

import (
	"context"
	"testing"

	"github.com/gravitee-io/terraform-provider-am/internal/planmodifiers/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestImmutable(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		state     types.String
		plan      types.String
		wantError string
	}{
		"create accepts any value": {
			state: types.StringNull(),
			plan:  types.StringValue("my-client"),
		},
		"unchanged value is accepted": {
			state: types.StringValue("my-client"),
			plan:  types.StringValue("my-client"),
		},
		"unknown plan is left as is": {
			state: types.StringValue("my-client"),
			plan:  types.StringUnknown(),
		},
		"changed value is rejected": {
			state:     types.StringValue("my-client"),
			plan:      types.StringValue("other-client"),
			wantError: "client_id cannot be changed once set",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				Path:       path.Root("client_id"),
				StateValue: tt.state,
				PlanValue:  tt.plan,
			}
			resp := &planmodifier.StringResponse{PlanValue: tt.plan}

			stringplanmodifier.Immutable().PlanModifyString(context.Background(), req, resp)

			if tt.wantError == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("expected no error, got: %s", resp.Diagnostics.Errors())
				}
				if !resp.PlanValue.Equal(tt.plan) {
					t.Fatalf("expected plan value %s to be left as is, got %s", tt.plan, resp.PlanValue)
				}
				return
			}

			if resp.Diagnostics.ErrorsCount() != 1 {
				t.Fatalf("expected 1 error, got: %s", resp.Diagnostics.Errors())
			}
			if got := resp.Diagnostics.Errors()[0].Detail(); got != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, got)
			}
		})
	}
}
