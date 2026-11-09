package examples_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const examplesPath = "../../examples"

var providerFile = filepath.Join("testdata", "provider.tf")

// The examples name each other by key (example-domain, signing-cert, mobile-app...), not by reference, so
// Terraform cannot order them: each stage only names resources of the stages before it.
var stages = [][]string{
	{"am_data_plane", "am_domain"},
	{"am_certificate", "am_identity_provider", "am_reporter"},
	{"am_application"},
	{"am_client_secret"},
}

// The first apply after importing these resources sends their configuration again: AM never returns the
// configuration of a data plane, and masks the secrets of the others (the mock does not).
var updatedOnImport = map[string]bool{"am_data_plane": true, "am_certificate": true, "am_identity_provider": true}

var (
	resourceAddress   = regexp.MustCompile(`(?m)^resource "(\w+)" "(\w+)"`)
	dataSourceAddress = regexp.MustCompile(`(?m)^data "(\w+)" "(\w+)"`)
	importCommand     = regexp.MustCompile(`^terraform import (\S+) '(.+)'`)
)

// Applies the resource, import and data-source examples as written, stage by stage, and checks every apply
// converges (plugin-testing fails a step whose next plan is not empty).
func TestResourceExamples(t *testing.T) {
	var steps []resource.TestStep
	for i := range stages {
		steps = append(steps, applyStep(t, resourceFiles(stages[:i+1]...)))
	}

	all := resourceFiles(stages...)
	steps = append(steps, dataSourcesStep(t, all))
	for _, resourceType := range flatten(stages) {
		steps = append(steps, importCommandStep(t, all, resourceType))
	}
	steps = append(steps, forgetStep(t), applyStep(t, with(all, importFiles())))

	// Destroys the stages in reverse order; the test case destroys the first one.
	for i := len(stages) - 1; i > 0; i-- {
		steps = append(steps, applyStep(t, resourceFiles(stages[:i]...)))
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps:    steps,
	})
}

// resourceFiles maps a config file name to the resource example of each resource type of the stages.
func resourceFiles(stages ...[]string) map[string]string {
	files := map[string]string{"provider.tf": providerFile}
	for _, resourceType := range flatten(stages) {
		files[resourceType+".tf"] = resourceExample(resourceType)
	}
	return files
}

func resourceExample(resourceType string) string {
	return filepath.Join(examplesPath, "resources", resourceType, "resource.tf")
}

func importFiles() map[string]string {
	files := map[string]string{}
	for _, resourceType := range flatten(stages) {
		files[resourceType+"_import.tf"] = filepath.Join(examplesPath, "resources", resourceType, "import-by-string-id.tf")
	}
	return files
}

// dataSourcesStep reads every data-source example and checks it finds the resource of the matching example.
func dataSourcesStep(t *testing.T, resources map[string]string) resource.TestStep {
	files := map[string]string{}
	var checks []resource.TestCheckFunc
	paths, _ := filepath.Glob(filepath.Join(examplesPath, "data-sources", "*", "data-source.tf"))
	for _, path := range paths {
		resourceType := filepath.Base(filepath.Dir(path))
		files[resourceType+"_data.tf"] = path
		checks = append(checks, resource.TestCheckResourceAttrPair(
			"data."+addresses(t, path, dataSourceAddress)[0], "name",
			addresses(t, resources[resourceType+".tf"], resourceAddress)[0], "name"))
	}
	step := applyStep(t, with(resources, files))
	step.Check = resource.ComposeAggregateTestCheckFunc(checks...)
	return step
}

// importCommandStep imports the resource with the ID of its import.sh example and checks the import plans no change.
func importCommandStep(t *testing.T, resources map[string]string, resourceType string) resource.TestStep {
	path := filepath.Join(examplesPath, "resources", resourceType, "import.sh")
	match := importCommand.FindStringSubmatch(strings.TrimSpace(readFile(t, path)))
	if match == nil {
		t.Fatalf("%s does not match %s", path, importCommand)
	}
	step := applyStep(t, resources)
	step.ResourceName = match[1]
	step.ImportState = true
	step.ImportStateKind = resource.ImportBlockWithID
	step.ImportStateId = match[2]
	step.ExpectNonEmptyPlan = updatedOnImport[resourceType]
	return step
}

// forgetStep removes every example resource from the state without destroying it, so the next step
// imports them back with the import-by-string-id.tf examples.
func forgetStep(t *testing.T) resource.TestStep {
	var removed strings.Builder
	for _, resourceType := range flatten(stages) {
		for _, address := range addresses(t, resourceExample(resourceType), resourceAddress) {
			fmt.Fprintf(&removed, "removed {\n  from = %s\n  lifecycle {\n    destroy = false\n  }\n}\n", address)
		}
	}
	removedFile := filepath.Join(t.TempDir(), "removed.tf")
	if err := os.WriteFile(removedFile, []byte(removed.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	return applyStep(t, map[string]string{"provider.tf": providerFile, "removed.tf": removedFile})
}

func applyStep(t *testing.T, files map[string]string) resource.TestStep {
	return resource.TestStep{
		ProtoV6ProviderFactories: testProviders(),
		ConfigDirectory:          config.StaticDirectory(configDir(t, files)),
		ConfigVariables:          exampleVariables(),
	}
}

// addresses returns the TYPE.NAME addresses of the blocks of the file matching pattern.
func addresses(t *testing.T, path string, pattern *regexp.Regexp) []string {
	var found []string
	for _, match := range pattern.FindAllStringSubmatch(readFile(t, path), -1) {
		found = append(found, match[1]+"."+match[2])
	}
	if len(found) == 0 {
		t.Fatalf("%s has no block matching %s", path, pattern)
	}
	return found
}

func readFile(t *testing.T, path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func with(files, more map[string]string) map[string]string {
	merged := map[string]string{}
	for _, m := range []map[string]string{files, more} {
		for name, path := range m {
			merged[name] = path
		}
	}
	return merged
}

func flatten(stages [][]string) []string {
	var all []string
	for _, stage := range stages {
		all = append(all, stage...)
	}
	return all
}
