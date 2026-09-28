package customtypes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuableWithSemanticEquals = (*MaskedJSON)(nil)
)

// mask is what AM returns in place of a sensitive plugin configuration value.
const mask = "********"

// maskOnly matches a value AM treats as masked: any run of asterisks.
var maskOnly = regexp.MustCompile(`^\*+$`)

// MaskedJSON is a JSON-encoded plugin configuration whose sensitive values AM returns as ********, including
// a password inside a connection URI. A secret changed outside Terraform cannot be detected.
type MaskedJSON struct {
	basetypes.StringValue
}

// Type returns a MaskedJSONType.
func (v MaskedJSON) Type(_ context.Context) attr.Type {
	return MaskedJSONType{}
}

// Equal returns true if the given value is equivalent.
func (v MaskedJSON) Equal(o attr.Value) bool {
	other, ok := o.(MaskedJSON)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true if the two configurations are equal as JSON once masked values are
// matched against whatever the other side holds at the same place. A value that is not valid JSON is compared
// as a plain string.
func (v MaskedJSON) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(MaskedJSON)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return v.StringValue.Equal(newValue.StringValue), diags
	}

	var current, next any
	if json.Unmarshal([]byte(v.ValueString()), &current) != nil || json.Unmarshal([]byte(newValue.ValueString()), &next) != nil {
		return v.ValueString() == newValue.ValueString(), diags
	}

	return maskedEqual(current, next), diags
}

func maskedEqual(a, b any) bool {
	switch aValue := a.(type) {
	case map[string]any:
		bValue, ok := b.(map[string]any)
		if !ok || len(aValue) != len(bValue) {
			return false
		}
		for key, aItem := range aValue {
			bItem, present := bValue[key]
			if !present || !maskedEqual(aItem, bItem) {
				return false
			}
		}
		return true
	case []any:
		bValue, ok := b.([]any)
		if !ok || len(aValue) != len(bValue) {
			return false
		}
		for i := range aValue {
			if !maskedEqual(aValue[i], bValue[i]) {
				return false
			}
		}
		return true
	case string:
		bValue, ok := b.(string)
		if !ok {
			return false
		}
		return aValue == bValue || matchesMasked(aValue, bValue) || matchesMasked(bValue, aValue)
	default:
		return a == b
	}
}

// matchesMasked reports whether value is what masked looks like before masking: masked is all asterisks, or
// every ******** in it stands for some run of characters in value (a password inside a URI).
func matchesMasked(masked, value string) bool {
	if maskOnly.MatchString(masked) {
		return true
	}
	if !strings.Contains(masked, mask) {
		return false
	}
	parts := strings.Split(masked, mask)
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	return regexp.MustCompile("^" + strings.Join(parts, ".*") + "$").MatchString(value)
}

// NewMaskedJSONNull creates a MaskedJSON with a null value.
func NewMaskedJSONNull() MaskedJSON {
	return MaskedJSON{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewMaskedJSONValue creates a MaskedJSON with a known value.
func NewMaskedJSONValue(value string) MaskedJSON {
	return MaskedJSON{
		StringValue: basetypes.NewStringValue(value),
	}
}
