package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

var getCmd = &cobra.Command{
	Use:   "get [KEY]",
	Short: "Get a config value or the entire config",
	Long:  `With no KEY, prints the full config as YAML. With KEY, prints that key's value.`,
	Example: `  stax config get
  stax config get api_url`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	v := viper.GetViper()
	if len(args) == 0 {
		settings := v.AllSettings()
		if len(settings) == 0 {
			fmt.Println("{}")
			return nil
		}
		bs, err := yaml.Marshal(settings)
		if err != nil {
			return err
		}
		fmt.Print(string(bs))
		return nil
	}
	key := args[0]
	if !v.IsSet(key) {
		fmt.Fprintln(os.Stderr, "not set")
		return nil
	}
	fmt.Println(v.Get(key))
	return nil
}
