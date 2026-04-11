package main

import (
	"os"

	"github.com/justn-hyeok/ganbatte/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
