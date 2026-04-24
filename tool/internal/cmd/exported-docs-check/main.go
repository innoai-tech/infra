package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/innoai-tech/infra/devpkg/exporteddoccheck"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}

	findings, err := exporteddoccheck.CheckPackages(filepath.Join(root, "pkg"))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}

	if len(findings) == 0 {
		return
	}

	for _, finding := range findings {
		_, _ = fmt.Fprintf(os.Stderr, "%s:%d missing doc for exported %s %s\n", finding.File, finding.Line, finding.Kind, finding.Name)
	}

	os.Exit(1)
}
