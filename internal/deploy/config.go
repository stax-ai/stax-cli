// Package deploy handles deployment configuration and orchestration for Kubernetes deployments.
package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config represents the deployment configuration from config.yaml.
type Config struct {
	ServiceName     string `yaml:"service_name"`
	PubSubTopic     string `yaml:"pubsub_topic"`
	Product         string `yaml:"product"`             // Required: "ta" or "cx"
	Environment     string `yaml:"environment"`         // Required: "prd" or "stg"
	Namespace       string `yaml:"namespace,omitempty"` // Overridden by product if not set
	ImageRegistry   string `yaml:"image_registry,omitempty"`
	ImageName       string `yaml:"image_name,omitempty"`
	ImageTag        string `yaml:"image_tag,omitempty"`
	RequireEmptyDir bool   `yaml:"require_empty_dir,omitempty"` // If true, mounts an emptyDir volume at /tmp
}

// Load reads and parses the deployment config from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}
	return &cfg, nil
}

// FindAndLoad searches for config.yaml starting at dir, then walking up to parent directories.
// If dir is empty, current working directory is used.
func FindAndLoadConfig(dir string) (*Config, string, error) {
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
		for _, name := range []string{"config.yaml", "config.yml"} {
			fpath := filepath.Join(p, name)
			if _, err := os.Stat(fpath); err != nil {
				continue
			}
			cfg, err := LoadConfig(fpath)
			if err != nil {
				return nil, "", err
			}
			return cfg, fpath, nil
		}
	}
	// try dir itself one more time with both names
	for _, name := range []string{"config.yaml", "config.yml"} {
		fpath := filepath.Join(abs, name)
		if _, err := os.Stat(fpath); err != nil {
			continue
		}
		cfg, err := LoadConfig(fpath)
		if err != nil {
			return nil, "", err
		}
		return cfg, fpath, nil
	}
	return nil, "", fmt.Errorf("deployment config.yaml not found in %s or parent directories", dir)
}

// Validate checks that required fields are set.
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if c.PubSubTopic == "" {
		return fmt.Errorf("pubsub_topic is required")
	}
	if c.Product == "" {
		return fmt.Errorf("product is required")
	}
	if c.Product != "ta" && c.Product != "cx" {
		return fmt.Errorf("product must be either 'ta' or 'cx', got: %s", c.Product)
	}
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if c.Environment != "prd" && c.Environment != "stg" {
		return fmt.Errorf("environment must be either 'prd' or 'stg', got: %s", c.Environment)
	}
	return nil
}

// GetImageName returns the image name, defaulting to service_name if not set.
func (c *Config) GetImageName() string {
	if c.ImageName != "" {
		return c.ImageName
	}
	return c.ServiceName
}

// GetImageTag returns the image tag, defaulting to "latest" if not set.
func (c *Config) GetImageTag() string {
	if c.ImageTag != "" {
		return c.ImageTag
	}
	return "latest"
}

// GetNamespace returns the namespace based on product, or the configured namespace if set.
func (c *Config) GetNamespace() string {
	if c.Namespace != "" {
		return c.Namespace
	}
	// Default namespace based on product
	if c.Product == "ta" {
		return "ta-automations"
	}
	if c.Product == "cx" {
		return "cx-automations"
	}
	return "default"
}

// GetBroker returns the broker name based on product.
func (c *Config) GetBroker() (string, error) {
	if c.Product == "ta" {
		return "ta-broker", nil
	}
	if c.Product == "cx" {
		return "cx-broker", nil
	}
	return "", fmt.Errorf("unknown product: %s", c.Product)
}

// GetFullImageName constructs the full image name from registry, name, and tag.
func (c *Config) GetFullImageName() (string, error) {
	if c.ImageRegistry == "" {
		return "", fmt.Errorf("image_registry is required to construct full image name")
	}
	registry := c.ImageRegistry
	if registry[len(registry)-1] == '/' {
		registry = registry[:len(registry)-1]
	}
	return fmt.Sprintf("%s/%s:%s", registry, c.GetImageName(), c.GetImageTag()), nil
}
