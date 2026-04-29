//go:build ignore
// +build ignore

// gendoc generates man pages for gnb commands.
// Usage: go run cmd/gendoc.go
package main

import (
	"fmt"
	"os"

	"github.com/bssm-oss/ganbatte/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "docs/man"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dir: %v\n", err)
		os.Exit(1)
	}

	header := &doc.GenManHeader{
		Title:   "GNB",
		Section: "1",
		Source:  "ganbatte",
		Manual:  "ganbatte Manual",
	}

	if err := doc.GenManTree(cmd.RootCmd, header, dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating man pages: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Man pages generated in %s/\n", dir)
}
