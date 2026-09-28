package acceptance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gravitee-io/terraform-provider-am/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const defaultServerURL = "http://localhost:8093/automation"

const mask = "********"

func TestMain(m *testing.M) {
	if os.Getenv("AM_SERVER_URL") == "" {
		os.Setenv("AM_SERVER_URL", defaultServerURL)
	}
	os.Exit(m.Run())
}

// testProviders serves the provider in-process for resource.TestCase.ProtoV6ProviderFactories.
func testProviders() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"am": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func preCheck(t *testing.T) {
	if os.Getenv("AM_SA_TOKEN") == "" {
		t.Fatal("AM_SA_TOKEN must hold an access token for the AM Automation API")
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

// uniqueKey returns a key valid for every Automation API resource.
func uniqueKey(prefix string) string {
	return prefix + "-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789")
}

// importStateIDFunc builds the JSON-encoded composite import ID from the resource's state attributes.
func importStateIDFunc(resourceAddress string, keys ...string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		attrs := s.RootModule().Resources[resourceAddress].Primary.Attributes
		id := make(map[string]string, len(keys))
		for _, key := range keys {
			id[key] = attrs[key]
		}
		b, err := json.Marshal(id)
		return string(b), err
	}
}

func automationPut(path string, body map[string]any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = automationCall(http.MethodPut, path, payload)
	return err
}

func automationGet(path string) (map[string]any, error) {
	body, err := automationCall(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	return out, json.Unmarshal(body, &out)
}

// automationCall calls the Automation API below /organizations/{orgId}/environments/{envId}, bypassing the provider.
func automationCall(method, path string, payload []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/organizations/%s/environments/%s%s",
		strings.TrimSuffix(os.Getenv("AM_SERVER_URL"), "/"), envOr("AM_ORG_ID", "DEFAULT"), envOr("AM_ENV_ID", "DEFAULT"), path)
	request, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+os.Getenv("AM_SA_TOKEN"))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s returned %d: %s", method, path, response.StatusCode, body)
	}
	return body, nil
}

// automationConfiguration returns the parsed `configuration` of a resource as the Automation API returns it.
func automationConfiguration(path string) (map[string]any, error) {
	body, err := automationGet(path)
	if err != nil {
		return nil, err
	}
	raw, ok := body["configuration"].(string)
	if !ok {
		return nil, fmt.Errorf("GET %s returned no configuration", path)
	}
	var configuration map[string]any
	return configuration, json.Unmarshal([]byte(raw), &configuration)
}

func domainConfig(key, name string) string {
	return fmt.Sprintf(`
resource "am_domain" "test" {
  key           = %[1]q
  name          = %[2]q
  path          = "/%[1]s"
  data_plane_id = "default"
}
`, key, name)
}
