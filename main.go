package main

import (
	"fmt"
	"github.com/miniscruff/changie/cmd"
	"os"
)

// goreleaser injected values
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	if err := cmd.Execute("v" + version); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
