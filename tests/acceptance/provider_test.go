package acceptance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-mock-server/server"
	"github.com/gravitee-io/terraform-provider-am/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const defaultServerURL = "http://localhost:8093/automation"

const mask = "********"

const mockTestEnv = "AM_MOCK"

func TestMain(m *testing.M) {
	if os.Getenv(mockTestEnv) == "true" {
		os.Exit(runOnMock(m))
	}
	if os.Getenv("AM_SERVER_URL") == "" {
		os.Setenv("AM_SERVER_URL", defaultServerURL)
	}
	os.Exit(m.Run())
}

// runOnMock runs the tests against an in-memory AM Automation API, closed once they are done.
func runOnMock(m *testing.M) int {
	mock := httptest.NewServer(server.New(server.NewMockAM()))
	defer mock.Close()
	os.Setenv("AM_SERVER_URL", mock.URL+server.BasePath)
	os.Setenv("AM_SA_TOKEN", "mock")
	return m.Run()
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

// domainVariables returns the variables of a testdata config holding the test domain, merged with the test's own.
func domainVariables(domainKey string, own config.Variables) config.Variables {
	variables := config.Variables{
		"domain_key":    config.StringVariable(domainKey),
		"data_plane_id": config.StringVariable(envOr("AM_TEST_DATA_PLANE_ID", "default")),
	}
	maps.Copy(variables, own)
	return variables
}

// checkAM fails unless the Automation API returns field as value for path.
func checkAM(path, field string, value any) resource.TestCheckFunc {
	return func(*terraform.State) error {
		body, err := automationGet(path)
		if err != nil {
			return err
		}
		if body[field] != value {
			return fmt.Errorf("AM returned %s of %s as %v instead of %v", field, path, body[field], value)
		}
		return nil
	}
}

// importStep imports address with the test's config and expects the state it had,
// except configuration, which AM masks or never returns.
func importStep(address, identifier string, variables config.Variables, idAttributes ...string) resource.TestStep {
	return resource.TestStep{
		ConfigDirectory:                      config.TestNameDirectory(),
		ConfigVariables:                      variables,
		ResourceName:                         address,
		ImportState:                          true,
		ImportStateIdFunc:                    importStateIDFunc(address, idAttributes...),
		ImportStateVerify:                    true,
		ImportStateVerifyIdentifierAttribute: identifier,
		ImportStateVerifyIgnore:              []string{"configuration"},
	}
}

// with returns a copy of variables where name is set to value.
func with(variables config.Variables, name string, value config.Variable) config.Variables {
	changed := maps.Clone(variables)
	changed[name] = value
	return changed
}

// immutableFieldSteps changes type, then system, expecting an in-place update AM rejects and the resource left intact.
// The test's config takes the resource's type and system from the "type" and "system" variables.
func immutableFieldSteps(address, path string, variables config.Variables, currentType, otherType string) []resource.TestStep {
	variables = with(variables, "type", config.StringVariable(currentType))
	updateRejected := func(variables config.Variables, message string) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
				},
			},
			ExpectError: regexp.MustCompile(message),
		}
	}
	return []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigVariables: variables,
		},
		updateRejected(with(variables, "type", config.StringVariable(otherType)), `The 'type' is immutable`),
		updateRejected(with(variables, "system", config.BoolVariable(true)), `The 'system' flag is immutable`),
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			Check: func(*terraform.State) error {
				body, err := automationGet(path)
				if err != nil {
					return err
				}
				if body["type"] != currentType || body["system"] == true {
					return fmt.Errorf("%s changed to type %v, system %v", path, body["type"], body["system"])
				}
				return nil
			},
		},
	}
}

func requiresLiveAM(t *testing.T) {
	if os.Getenv(mockTestEnv) == "true" {
		t.Skip("Acceptance tests skipped as it requires a live AM instance")
	}
}
