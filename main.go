package main

import "github.com/miniscruff/changie/cmd"

// goreleaser injected values
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cmd.Execute("v" + version)
}
