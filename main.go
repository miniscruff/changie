package main

import (
	"fmt"
	"os"

	"github.com/miniscruff/changie/cmd"
)

// goreleaser injected values
var version = "dev"

func main() {
	rootCmd := cmd.RootCmd()
	rootCmd.Version = "v" + version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
