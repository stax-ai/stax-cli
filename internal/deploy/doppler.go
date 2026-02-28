package deploy

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// DopplerSecret represents a Doppler Secret CRD manifest.
type DopplerSecret struct {
	APIVersion string                `yaml:"apiVersion"`
	Kind       string                `yaml:"kind"`
	Metadata   DopplerSecretMetadata `yaml:"metadata"`
	Spec       DopplerSecretSpec     `yaml:"spec"`
}

// DopplerSecretMetadata contains metadata for the DopplerSecret.
type DopplerSecretMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// DopplerSecretSpec contains the spec for the DopplerSecret.
type DopplerSecretSpec struct {
	TokenSecret   DopplerTokenSecret   `yaml:"tokenSecret"`
	Project       string               `yaml:"project"`
	Config        string               `yaml:"config"`
	ManagedSecret DopplerManagedSecret `yaml:"managedSecret"`
}

// DopplerTokenSecret contains the token secret reference.
type DopplerTokenSecret struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// DopplerManagedSecret contains the managed secret configuration.
type DopplerManagedSecret struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Type      string `yaml:"type"`
}

// GenerateDopplerSecret generates a DopplerSecret YAML manifest.
func GenerateDopplerSecret(cfg *Config, targetNamespace string) (*DopplerSecret, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if cfg.Environment == "" {
		return nil, fmt.Errorf("environment is required")
	}
	if cfg.Environment != "prd" && cfg.Environment != "stg" {
		return nil, fmt.Errorf("environment must be either 'prd' or 'stg', got: %s", cfg.Environment)
	}

	// DopplerSecret name follows pattern: dopplersecret-{service_name}
	dopplerSecretName := fmt.Sprintf("dopplersecret-%s", cfg.ServiceName)
	// Managed secret name matches the DopplerSecret name
	managedSecretName := fmt.Sprintf("secret-%s", cfg.ServiceName)
	// Token secret name follows pattern: {service_name}-token-secret
	tokenSecretName := fmt.Sprintf("%s-token-secret", cfg.ServiceName)

	ds := &DopplerSecret{
		APIVersion: "secrets.doppler.com/v1alpha1",
		Kind:       "DopplerSecret",
		Metadata: DopplerSecretMetadata{
			Name:      dopplerSecretName,
			Namespace: "doppler-operator-system", // Fixed namespace for Doppler operator
		},
		Spec: DopplerSecretSpec{
			TokenSecret: DopplerTokenSecret{
				Name:      tokenSecretName,
				Namespace: "doppler-operator-system",
			},
			Project: cfg.ServiceName, // Service name as Doppler project
			Config:  cfg.Environment, // Environment (prd or stg) as Doppler config
			ManagedSecret: DopplerManagedSecret{
				Name:      managedSecretName,
				Namespace: targetNamespace, // Target namespace where the secret will be created
				Type:      "Opaque",
			},
		},
	}

	return ds, nil
}

// ToYAML converts the DopplerSecret to YAML.
func (ds *DopplerSecret) ToYAML() ([]byte, error) {
	return yaml.Marshal(ds)
}
