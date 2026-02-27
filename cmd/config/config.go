// Package config provides the "stax config" command for global CLI settings (.stax.yml).
package config

import (
	"github.com/spf13/cobra"
)

// Cmd is the "stax config" command. It reads or writes global CLI config (get/set).
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Read or write stax CLI configuration",
	Long: `Manage global stax-cli configuration stored in .stax.yml (or .stax.yaml).
Config is searched in: user config dir, home, then current directory. Use the
global --config-file flag to point to a specific file.

Subcommands:
  get   Print a config key or the entire config (YAML)
  set   Set a config key to a value`,
}

func init() {
	Cmd.AddCommand(getCmd, setCmd)
}
