package workrun_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/innoai-tech/infra/devpkg/workrun"
)

func TestRunnerContext(t *testing.T) {
	t.Run("GIVEN a context with specific dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "example")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origWd)

		t.Run("WHEN context is ./example and args contain {} THEN {} is replaced with ./", func(t *testing.T) {
			r := &workrun.Runner{}
			r.SetDefaults()
			r.Context = "./example"
			r.Args = []string{"echo", "{}"}

			if err := r.Run(context.Background()); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("WHEN context is ./example/... and args contain {} THEN {} is replaced with ./...", func(t *testing.T) {
			r := &workrun.Runner{}
			r.SetDefaults()
			r.Context = "./example/..."
			r.Args = []string{"echo", "{}"}

			if err := r.Run(context.Background()); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("GIVEN context is ./...", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "example")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		workContent := []byte("go 1.26.5\n\nuse ./example\n")
		if err := os.WriteFile(filepath.Join(tmpDir, "go.work"), workContent, 0644); err != nil {
			t.Fatal(err)
		}

		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origWd)

		t.Run("WHEN args contain {} THEN {} is replaced with ./... for each module", func(t *testing.T) {
			r := &workrun.Runner{}
			r.SetDefaults()
			r.Args = []string{"echo", "{}"}

			if err := r.Run(context.Background()); err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestRunnerDefaultContext(t *testing.T) {
	t.Run("GIVEN a Runner with default context", func(t *testing.T) {
		r := &workrun.Runner{}

		t.Run("WHEN SetDefaults is called THEN Context is ./...", func(t *testing.T) {
			r.SetDefaults()
			Then(t, "context defaults to ./...",
				Expect(r.Context, Equal("./...")),
			)
		})

		t.Run("WHEN Context is already set THEN SetDefaults does not overwrite", func(t *testing.T) {
			r.Context = "./custom"
			r.SetDefaults()
			Then(t, "context remains ./custom",
				Expect(r.Context, Equal("./custom")),
			)
		})
	})
}
