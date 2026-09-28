package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestMaskedJSON_StringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		configured string
		returned   string
		expected   bool
	}{
		"identical": {
			configured: `{"clientId":"app","clientSecret":"s3cr3t"}`,
			returned:   `{"clientId":"app","clientSecret":"s3cr3t"}`,
			expected:   true,
		},
		"key order and whitespace": {
			configured: `{"clientId":"app", "clientSecret":"s3cr3t"}`,
			returned:   `{"clientSecret":"s3cr3t","clientId":"app"}`,
			expected:   true,
		},
		"masked secret": {
			configured: `{"clientId":"app","clientSecret":"s3cr3t"}`,
			returned:   `{"clientId":"app","clientSecret":"********"}`,
			expected:   true,
		},
		"masked secret in an array": {
			configured: `{"users":[{"username":"alice","password":"a"},{"username":"bob","password":"b"}]}`,
			returned:   `{"users":[{"username":"alice","password":"********"},{"username":"bob","password":"********"}]}`,
			expected:   true,
		},
		"masked password inside a uri": {
			configured: `{"uri":"mongodb://am:s3cr3t@mongo:27017/am"}`,
			returned:   `{"uri":"mongodb://am:********@mongo:27017/am"}`,
			expected:   true,
		},
		"masked object value": {
			configured: `{"jks":"{\"name\":\"keystore.jks\",\"content\":\"AAAA\"}","alias":"am"}`,
			returned:   `{"jks":"********","alias":"am"}`,
			expected:   true,
		},
		"non-secret change": {
			configured: `{"clientId":"app","clientSecret":"s3cr3t"}`,
			returned:   `{"clientId":"other","clientSecret":"********"}`,
			expected:   false,
		},
		"uri host change around a masked password": {
			configured: `{"uri":"mongodb://am:s3cr3t@mongo:27017/am"}`,
			returned:   `{"uri":"mongodb://am:********@other:27017/am"}`,
			expected:   false,
		},
		"field removed": {
			configured: `{"clientId":"app","scope":"openid"}`,
			returned:   `{"clientId":"app"}`,
			expected:   false,
		},
		"array length differs": {
			configured: `{"users":[{"username":"alice"}]}`,
			returned:   `{"users":[{"username":"alice"},{"username":"bob"}]}`,
			expected:   false,
		},
		"number and string differ": {
			configured: `{"port":389}`,
			returned:   `{"port":"389"}`,
			expected:   false,
		},
		"a mask never stands for an object": {
			configured: `{"ldapConfig":{"password":"s3cr3t"}}`,
			returned:   `{"ldapConfig":"********"}`,
			expected:   false,
		},
		"not json, equal": {
			configured: `not json`,
			returned:   `not json`,
			expected:   true,
		},
		"not json, different": {
			configured: `not json`,
			returned:   `********`,
			expected:   false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			configured := NewMaskedJSONValue(testCase.configured)
			returned := NewMaskedJSONValue(testCase.returned)

			forward, diags := configured.StringSemanticEquals(context.Background(), returned)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			backward, diags := returned.StringSemanticEquals(context.Background(), configured)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if forward != testCase.expected || backward != testCase.expected {
				t.Errorf("expected %t both ways, got %t forward and %t backward", testCase.expected, forward, backward)
			}
		})
	}
}

func TestMaskedJSON_StringSemanticEquals_WrongType(t *testing.T) {
	t.Parallel()

	_, diags := NewMaskedJSONValue(`{}`).StringSemanticEquals(context.Background(), basetypes.NewStringValue(`{}`))

	if !diags.HasError() {
		t.Error("expected an error diagnostic for a value of another type")
	}
}

func TestMaskedJSONType_ValueFromString(t *testing.T) {
	t.Parallel()

	valuable, diags := MaskedJSONType{}.ValueFromString(context.Background(), basetypes.NewStringPointerValue(nil))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	value, ok := valuable.(MaskedJSON)
	if !ok {
		t.Fatalf("expected a MaskedJSON, got %T", valuable)
	}
	if !value.IsNull() {
		t.Error("expected a null value for a nil string")
	}
}
