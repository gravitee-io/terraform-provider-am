// Package examples_test applies the documented examples as they are written: the resource, import and
// data-source examples of the registry docs, and the use-cases the guides are built from.
package examples_test

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-mock-server/server"
	"github.com/gravitee-io/terraform-provider-am/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

const defaultServerURL = "http://localhost:8093/automation"

const mockTestEnv = "AM_MOCK"

func TestMain(m *testing.M) {
	// The examples declare source = "gravitee-io/am": serve the in-process provider under that address.
	os.Setenv("TF_ACC_PROVIDER_NAMESPACE", "gravitee-io")
	if os.Getenv(mockTestEnv) == "true" {
		os.Exit(runOnMock(m))
	}
	if os.Getenv("AM_SERVER_URL") == "" {
		os.Setenv("AM_SERVER_URL", defaultServerURL)
	}
	os.Exit(m.Run())
}

func runOnMock(m *testing.M) int {
	mock := httptest.NewServer(server.New(server.NewMockAM()))
	defer mock.Close()
	os.Setenv("AM_SERVER_URL", mock.URL+server.BasePath)
	os.Setenv("AM_SA_TOKEN", "mock")
	return m.Run()
}

func preCheck(t *testing.T) {
	if os.Getenv("AM_SA_TOKEN") == "" {
		t.Fatal("AM_SA_TOKEN must hold an access token for the AM Automation API")
	}
}

func testProviders() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"am": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

// exampleVariables holds a value for every variable an example declares; an example ignores the others.
// The keystore passwords match testdata/keystore.jks, a key pair generated with:
//
//	keytool -genkeypair -alias signing -keyalg RSA -keysize 2048 -validity 3650 -dname CN=example \
//	  -storetype JKS -keystore keystore.jks -storepass storepass123 -keypass keypass123
func exampleVariables() config.Variables {
	return config.Variables{
		"keystore_password":      config.StringVariable("storepass123"),
		"key_password":           config.StringVariable("keypass123"),
		"keystore_2026_password": config.StringVariable("storepass123"),
		"key_2026_password":      config.StringVariable("keypass123"),
		"keystore_2027_password": config.StringVariable("storepass123"),
		"key_2027_password":      config.StringVariable("keypass123"),
		"alice_password":         config.StringVariable("Alice-Example-1"),
		"client_secret":          config.StringVariable(acctest.RandString(64)),
		"next_client_secret":     config.StringVariable(acctest.RandString(64)),
		// The MongoDB AM reaches, as for the acceptance tests.
		"mongodb_host":            config.StringVariable(envOr("AM_TEST_MONGODB_HOST", "mongodb")),
		"mongodb_port":            config.StringVariable(envOr("AM_TEST_MONGODB_PORT", "27017")),
		"corporate_client_secret": config.StringVariable(acctest.RandString(32)),
		"kafka_password":          config.StringVariable(acctest.RandString(32)),
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

// keystores are the files the certificate examples read from ${path.module}, all copies of the test keystore.
var keystores = map[string]string{
	"keystore.jks":      filepath.Join("testdata", "keystore.jks"),
	"keystore-2026.jks": filepath.Join("testdata", "keystore.jks"),
	"keystore-2027.jks": filepath.Join("testdata", "keystore.jks"),
}

// configDir returns a new directory holding the given files, named by the map keys, and the keystores.
func configDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, source := range with(files, keystores) {
		content, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
