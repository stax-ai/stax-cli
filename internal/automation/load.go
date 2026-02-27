package automation

import (
	"os"
	"path/filepath"
)

const (
	configFilename = "config.yml"
	configAltName  = "config.yaml"
)

// Load reads and parses the automation config from the given path.
// The path must point to a config.yml (or config.yaml) file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := unmarshalYAML(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// FindAndLoad searches for config.yml (or config.yaml) starting at dir,
// then walking up to parent directories, and loads the first one found.
// If dir is empty, current working directory is used.
func FindAndLoad(dir string) (*Config, string, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, "", err
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}
	for p := abs; p != filepath.Dir(p); p = filepath.Dir(p) {
		for _, name := range []string{configFilename, configAltName} {
			fpath := filepath.Join(p, name)
			if _, err := os.Stat(fpath); err != nil {
				continue
			}
			cfg, err := Load(fpath)
			if err != nil {
				return nil, "", err
			}
			return cfg, fpath, nil
		}
	}
	// try dir itself one more time with both names
	for _, name := range []string{configFilename, configAltName} {
		fpath := filepath.Join(abs, name)
		if _, err := os.Stat(fpath); err != nil {
			continue
		}
		cfg, err := Load(fpath)
		if err != nil {
			return nil, "", err
		}
		return cfg, fpath, nil
	}
	return nil, "", os.ErrNotExist
}
