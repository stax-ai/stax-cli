package deploy

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	deployDirName = "deploy"
	knativeServiceFile = "knative-service.yaml"
	pipeFile           = "pipe.yaml"
	integrationFile    = "integration.yaml"
	dopplerSecretFile  = "doppler-secret.yaml"
)

// MemoizeResources saves all generated YAML resources to the deploy/ folder.
func MemoizeResources(projectRoot string, knativeService *KnativeService, pipe *Pipe, integration *Integration, dopplerSecret *DopplerSecret) error {
	deployDir := filepath.Join(projectRoot, deployDirName)

	// Create deploy directory if it doesn't exist
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return fmt.Errorf("create deploy directory: %w", err)
	}

	// Save Knative Service
	if knativeService != nil {
		knativeYAML, err := knativeService.ToYAML()
		if err != nil {
			return fmt.Errorf("marshal knative service: %w", err)
		}
		knativePath := filepath.Join(deployDir, knativeServiceFile)
		if err := os.WriteFile(knativePath, knativeYAML, 0644); err != nil {
			return fmt.Errorf("write knative service: %w", err)
		}
	}

	// Save Pipe
	if pipe != nil {
		pipeYAML, err := pipe.ToYAML()
		if err != nil {
			return fmt.Errorf("marshal pipe: %w", err)
		}
		pipePath := filepath.Join(deployDir, pipeFile)
		if err := os.WriteFile(pipePath, pipeYAML, 0644); err != nil {
			return fmt.Errorf("write pipe: %w", err)
		}
	}

	// Save Integration
	if integration != nil {
		integrationYAML, err := integration.ToYAML()
		if err != nil {
			return fmt.Errorf("marshal integration: %w", err)
		}
		integrationPath := filepath.Join(deployDir, integrationFile)
		if err := os.WriteFile(integrationPath, integrationYAML, 0644); err != nil {
			return fmt.Errorf("write integration: %w", err)
		}
	}

	// Save DopplerSecret
	if dopplerSecret != nil {
		dopplerYAML, err := dopplerSecret.ToYAML()
		if err != nil {
			return fmt.Errorf("marshal doppler secret: %w", err)
		}
		dopplerPath := filepath.Join(deployDir, dopplerSecretFile)
		if err := os.WriteFile(dopplerPath, dopplerYAML, 0644); err != nil {
			return fmt.Errorf("write doppler secret: %w", err)
		}
	}

	return nil
}

// GetDeployDir returns the path to the deploy directory.
func GetDeployDir(projectRoot string) string {
	return filepath.Join(projectRoot, deployDirName)
}

// GetKnativeServicePath returns the path to the knative service YAML file.
func GetKnativeServicePath(projectRoot string) string {
	return filepath.Join(GetDeployDir(projectRoot), knativeServiceFile)
}

// GetPipePath returns the path to the pipe YAML file.
func GetPipePath(projectRoot string) string {
	return filepath.Join(GetDeployDir(projectRoot), pipeFile)
}

// GetIntegrationPath returns the path to the integration YAML file.
func GetIntegrationPath(projectRoot string) string {
	return filepath.Join(GetDeployDir(projectRoot), integrationFile)
}

// GetDopplerSecretPath returns the path to the doppler secret YAML file.
func GetDopplerSecretPath(projectRoot string) string {
	return filepath.Join(GetDeployDir(projectRoot), dopplerSecretFile)
}
