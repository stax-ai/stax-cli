/*
Copyright © 2026 Stax.ai, Inc. <developers@stax.ai>

*/
package main

import "github.com/stax-ai/stax-cli/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	cmd.SetVersion(version)
}

func main() {
	cmd.Execute()
}
