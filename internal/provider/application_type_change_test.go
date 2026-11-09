package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIsConfigured(t *testing.T) {
	t.Parallel()

	oauthType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"grant_types": tftypes.List{ElementType: tftypes.String}}}
	settingsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"oauth": oauthType}}
	configType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"settings": settingsType}}
	config := func(settings tftypes.Value) tftypes.Value {
		return tftypes.NewValue(configType, map[string]tftypes.Value{"settings": settings})
	}
	oauth := func(grantTypes tftypes.Value) tftypes.Value {
		return tftypes.NewValue(settingsType, map[string]tftypes.Value{
			"oauth": tftypes.NewValue(oauthType, map[string]tftypes.Value{"grant_types": grantTypes}),
		})
	}
	listType := tftypes.List{ElementType: tftypes.String}

	tests := map[string]struct {
		config tftypes.Value
		want   bool
	}{
		"declared": {
			config: config(oauth(tftypes.NewValue(listType, []tftypes.Value{tftypes.NewValue(tftypes.String, "client_credentials")}))),
			want:   true,
		},
		"declared from a value known after apply": {
			config: config(oauth(tftypes.NewValue(listType, tftypes.UnknownValue))),
			want:   true,
		},
		"left unset": {
			config: config(oauth(tftypes.NewValue(listType, nil))),
			want:   false,
		},
		"no settings block": {
			config: config(tftypes.NewValue(settingsType, nil)),
			want:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := isConfigured(tt.config, "settings", "oauth", "grant_types"); got != tt.want {
				t.Fatalf("isConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}
