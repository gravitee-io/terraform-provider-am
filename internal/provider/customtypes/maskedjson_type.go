package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*MaskedJSONType)(nil)
)

// MaskedJSONType is the attribute type of MaskedJSON.
type MaskedJSONType struct {
	basetypes.StringType
}

// String returns a human-readable string of the type name.
func (t MaskedJSONType) String() string {
	return "customtypes.MaskedJSONType"
}

// ValueType returns the Value type.
func (t MaskedJSONType) ValueType(ctx context.Context) attr.Value {
	return MaskedJSON{}
}

// Equal returns true if the given type is equivalent.
func (t MaskedJSONType) Equal(o attr.Type) bool {
	other, ok := o.(MaskedJSONType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t MaskedJSONType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return MaskedJSON{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t MaskedJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
