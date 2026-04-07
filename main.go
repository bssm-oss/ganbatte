package main

import (
	"os"

	"github.com/bssm-oss/ganbatte/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
