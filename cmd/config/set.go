package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalconfig "github.com/stax-ai/stax-cli/internal/config"
)

var setCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set a config value",
	Long:  `Writes KEY=VALUE to the global config file. Creates the file if it does not exist.`,
	Example: `  stax config set api_url https://api.example.com
  stax config set profile default`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func runSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]
	v := viper.GetViper()
	v.Set(key, value)

	cfgPath := v.ConfigFileUsed()
	if cfgPath == "" {
		cfgPath = internalconfig.DefaultWritePath()
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
		v.SetConfigFile(cfgPath)
	}
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
