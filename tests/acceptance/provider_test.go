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
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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

// checkAMNested fails unless the Automation API returns value at the nested fields of path
// (a list index is a field too, e.g. "redirectUris", "0").
func checkAMNested(path string, value any, fields ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		body, err := automationGet(path)
		if err != nil {
			return err
		}
		var got any = body
		for _, field := range fields {
			switch node := got.(type) {
			case map[string]any:
				got = node[field]
			case []any:
				index, err := strconv.Atoi(field)
				if err != nil || index >= len(node) {
					return fmt.Errorf("AM returned no %s in %s", strings.Join(fields, "."), path)
				}
				got = node[index]
			default:
				return fmt.Errorf("AM returned no %s in %s", strings.Join(fields, "."), path)
			}
		}
		if got != value {
			return fmt.Errorf("AM returned %s of %s as %v instead of %v", strings.Join(fields, "."), path, got, value)
		}
		return nil
	}
}

// checkAMMatchesState fails unless the Automation API returns field of path as the state holds attribute of address.
// Timestamps match when they are the same instant, whatever their RFC 3339 precision.
func checkAMMatchesState(path, field, address, attribute string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		body, err := automationGet(path)
		if err != nil {
			return err
		}
		got, _ := body[field].(string)
		want := s.RootModule().Resources[address].Primary.Attributes[attribute]
		if got == want || sameInstant(got, want) {
			return nil
		}
		return fmt.Errorf("AM returned %s of %s as %q, state holds %q", field, path, got, want)
	}
}

func sameInstant(a, b string) bool {
	ta, errA := time.Parse(time.RFC3339Nano, a)
	tb, errB := time.Parse(time.RFC3339Nano, b)
	return errA == nil && errB == nil && ta.Equal(tb)
}

// checkOutputMatchesAttribute fails unless the output holds the state value of attribute of address.
func checkOutputMatchesAttribute(output, address, attribute string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		want := s.RootModule().Resources[address].Primary.Attributes[attribute]
		return resource.TestCheckOutput(output, want)(s)
	}
}

// conflictingMock serves a fresh AM mock that answers the first method request on a path ending with suffix with 409 Conflict,
// as AM does when concurrent writes on an application keep colliding. It returns the mock's URL and the 409 count.
// The test needs no live AM, so it is skipped unless the tests run on the mock.
func conflictingMock(t *testing.T, method, suffix string) (string, *atomic.Int32) {
	if os.Getenv(mockTestEnv) != "true" {
		t.Skip("409 retries are tested on the AM mock only")
	}
	mock := server.New(server.NewMockAM())
	var conflicts atomic.Int32
	conflicting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method && strings.HasSuffix(r.URL.Path, suffix) && conflicts.CompareAndSwap(0, 1) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"message":"Concurrent modification, retry the request","http_status":409}`)
			return
		}
		mock.ServeHTTP(w, r)
	}))
	t.Cleanup(conflicting.Close)
	return conflicting.URL + server.BasePath, &conflicts
}

func checkConflictsAnswered(conflicts *atomic.Int32, want int32) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := conflicts.Load(); got != want {
			return fmt.Errorf("the mock answered %d requests with 409 instead of %d", got, want)
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

// onLiveAM runs check on a live AM only, for a behaviour the mock does not reproduce.
func onLiveAM(check resource.TestCheckFunc) resource.TestCheckFunc {
	if os.Getenv(mockTestEnv) == "true" {
		return func(*terraform.State) error { return nil }
	}
	return check
}

func requiresLiveAM(t *testing.T) {
	if os.Getenv(mockTestEnv) == "true" {
		t.Skip("Acceptance tests skipped as it requires a live AM instance")
	}
}
