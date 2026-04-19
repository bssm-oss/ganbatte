package main

import (
	"os"

	"github.com/justn-hyeok/ganbatte/cmd"
)

func main() {
	cmd.RootCmd.SetOut(os.Stdout)
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
