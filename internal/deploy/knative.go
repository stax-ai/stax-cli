package deploy

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// KnativeService represents a Knative Service manifest.
type KnativeService struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   KnativeServiceMetadata `yaml:"metadata"`
	Spec       KnativeServiceSpec     `yaml:"spec"`
}

// KnativeServiceMetadata contains metadata for the Knative Service.
type KnativeServiceMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// KnativeServiceSpec contains the spec for the Knative Service.
type KnativeServiceSpec struct {
	Template KnativeServiceTemplate `yaml:"template"`
}

// KnativeServiceTemplate contains the template for the Knative Service.
type KnativeServiceTemplate struct {
	Metadata KnativeServiceTemplateMetadata `yaml:"metadata"`
	Spec     KnativeServiceTemplateSpec     `yaml:"spec"`
}

// KnativeServiceTemplateMetadata contains template metadata.
type KnativeServiceTemplateMetadata struct {
	Annotations map[string]string `yaml:"annotations"`
}

// KnativeServiceTemplateSpec contains the template spec.
type KnativeServiceTemplateSpec struct {
	ContainerConcurrency int64                     `yaml:"containerConcurrency"`
	Containers           []KnativeServiceContainer `yaml:"containers"`
	Volumes              []map[string]interface{}  `yaml:"volumes,omitempty"`
}

// KnativeServiceContainer represents a container in the Knative Service.
type KnativeServiceContainer struct {
	Image           string                   `yaml:"image"`
	Ports           []KnativeServicePort     `yaml:"ports,omitempty"`
	Env             []map[string]interface{} `yaml:"env,omitempty"`
	EnvFrom         []map[string]interface{} `yaml:"envFrom,omitempty"`
	VolumeMounts    []map[string]interface{} `yaml:"volumeMounts,omitempty"`
	SecurityContext map[string]interface{}   `yaml:"securityContext,omitempty"`
}

// KnativeServicePort represents a port in the container.
type KnativeServicePort struct {
	ContainerPort int    `yaml:"containerPort"`
	Name          string `yaml:"name,omitempty"`
}

// GenerateKnativeService generates a Knative Service YAML manifest.
func GenerateKnativeService(cfg *Config, image string) (*KnativeService, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if image == "" {
		return nil, fmt.Errorf("image is required")
	}

	// Managed secret name matches the DopplerSecret name: dopplersecret-{service_name}
	secretName := fmt.Sprintf("secret-%s", cfg.ServiceName)

	// Build container spec
	container := KnativeServiceContainer{
		Image: image,
		Ports: []KnativeServicePort{
			{
				ContainerPort: 8080,
				Name:          "http1",
			},
		},
		// Reference Doppler secret via envFrom
		EnvFrom: []map[string]interface{}{
			{
				"secretRef": map[string]interface{}{
					"name": secretName,
				},
			},
		},
		// Security context with secure defaults
		// Note: runAsNonRoot is not set because the image uses a non-numeric user (cnb)
		// which Kubernetes cannot verify as non-root. The cnb user from buildpacks is non-root.
		SecurityContext: map[string]interface{}{
			"allowPrivilegeEscalation": false,
			"capabilities": map[string]interface{}{
				"drop": []string{"ALL"},
			},
			"seccompProfile": map[string]interface{}{
				"type": "RuntimeDefault",
			},
		},
	}

	// Add emptyDir volume if required
	var volumes []map[string]interface{}
	if cfg.RequireEmptyDir {
		volumes = []map[string]interface{}{
			{
				"name":     "tmp",
				"emptyDir": map[string]interface{}{},
			},
		}
		container.VolumeMounts = []map[string]interface{}{
			{
				"name":      "tmp",
				"mountPath": "/tmp",
			},
		}
	}

	// Build Datadog labels and annotations
	// Container name for Datadog annotations (default container name in Knative)
	containerName := "user-container"

	// Service-level labels for Datadog
	labels := map[string]string{
		"app": cfg.ServiceName,
	}

	// Add Datadog tags as labels
	labels["tags.datadoghq.com/env"] = cfg.Environment
	labels["tags.datadoghq.com/service"] = cfg.ServiceName
	if cfg.ImageTag != "" {
		labels["tags.datadoghq.com/version"] = cfg.ImageTag
	}

	// Template-level annotations for Datadog injection
	annotations := map[string]string{
		"autoscaling.knative.dev/minScale": "0",
		"autoscaling.knative.dev/maxScale": "10",
	}

	// Datadog auto-injection annotations
	annotations[fmt.Sprintf("ad.datadoghq.com/%s.logs", containerName)] = fmt.Sprintf(`[{"source":"%s","service":"%s"}]`, cfg.ServiceName, cfg.ServiceName)
	annotations[fmt.Sprintf("ad.datadoghq.com/%s.env", containerName)] = cfg.Environment
	annotations[fmt.Sprintf("ad.datadoghq.com/%s.service", containerName)] = cfg.ServiceName
	if cfg.ImageTag != "" {
		annotations[fmt.Sprintf("ad.datadoghq.com/%s.version", containerName)] = cfg.ImageTag
	}

	svc := &KnativeService{
		APIVersion: "serving.knative.dev/v1",
		Kind:       "Service",
		Metadata: KnativeServiceMetadata{
			Name:      cfg.ServiceName,
			Namespace: cfg.GetNamespace(),
			Labels:    labels,
		},
		Spec: KnativeServiceSpec{
			Template: KnativeServiceTemplate{
				Metadata: KnativeServiceTemplateMetadata{
					Annotations: annotations,
				},
				Spec: KnativeServiceTemplateSpec{
					ContainerConcurrency: 1, // Only 1 request at a time
					Containers: []KnativeServiceContainer{
						container,
					},
					Volumes: volumes,
				},
			},
		},
	}

	return svc, nil
}

// ToYAML converts the Knative Service to YAML.
func (ks *KnativeService) ToYAML() ([]byte, error) {
	return yaml.Marshal(ks)
}
