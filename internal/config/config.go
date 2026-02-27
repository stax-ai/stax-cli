// Package config manages global stax-cli configuration (.stax.yml). It is loaded at
// startup via Init; cmd/config uses it for get/set. Search order: --config-file,
// then user config dir, home, current directory.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configName = ".stax"
	configType = "yaml"
)

// Config holds the global CLI configuration. Viper unmarshals into this struct.
// Use mapstructure tags for YAML keys (e.g. api_url, profile).
type Config struct {
	APIURL  string         `mapstructure:"api_url"`
	Profile string         `mapstructure:"profile"`
	Extra   map[string]any `mapstructure:",remain"`
}

// Init loads configuration from cfgFile if set, otherwise searches for
// .stax.yml or .stax.yaml in APPDIR (UserConfigDir), home, then current dir.
// Uses the global viper instance. Missing config file is not an error.
func Init(cfgFile string) error {
	v := viper.GetViper()
	v.SetConfigType(configType)

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		return v.Unmarshal(&Config{})
	}

	searchDirs := searchDirs()
	extensions := []string{".yml", ".yaml"}

	for _, dir := range searchDirs {
		for _, ext := range extensions {
			path := filepath.Join(dir, configName+ext)
			if _, err := os.Stat(path); err == nil {
				v.SetConfigFile(path)
				if err := v.ReadInConfig(); err != nil {
					return err
				}
				return v.Unmarshal(&Config{})
			}
		}
	}

	return nil
}

// searchDirs returns config search directories in order: APPDIR, home, current.
func searchDirs() []string {
	dirs := make([]string, 0, 3)
	if d, err := os.UserConfigDir(); err == nil && d != "" {
		dirs = append(dirs, d)
	}
	if d, err := os.UserHomeDir(); err == nil && d != "" {
		dirs = append(dirs, d)
	}
	dirs = append(dirs, ".")
	return dirs
}

// Get unmarshals the current viper state into a Config and returns it.
func Get() (*Config, error) {
	var c Config
	if err := viper.GetViper().Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// DefaultWritePath returns the path to use when writing config and no file was loaded.
// Caller should ensure the parent directory exists (e.g. os.MkdirAll(filepath.Dir(path), 0755)).
func DefaultWritePath() string {
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, configName+".yaml")
}
