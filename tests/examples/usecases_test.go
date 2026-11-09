package examples_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Applies each use-case the guides are built from, checks the next plan is empty, and destroys it.
func TestUseCases(t *testing.T) {
	dirs, err := os.ReadDir(filepath.Join(examplesPath, "use-cases"))
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		t.Run(dir.Name(), func(t *testing.T) {
			t.Parallel()
			resource.Test(t, resource.TestCase{
				PreCheck: func() { preCheck(t) },
				Steps:    []resource.TestStep{applyStep(t, useCaseFiles(t, dir.Name()))},
			})
		})
	}
}

func useCaseFiles(t *testing.T, name string) map[string]string {
	paths, err := filepath.Glob(filepath.Join(examplesPath, "use-cases", name, "*.tf"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, path := range paths {
		files[filepath.Base(path)] = path
	}
	return files
}
