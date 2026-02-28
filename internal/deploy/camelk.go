package deploy

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Pipe represents a camel-k Pipe manifest.
type Pipe struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   PipeMetadata `yaml:"metadata"`
	Spec       PipeSpec     `yaml:"spec"`
}

// PipeMetadata contains metadata for the Pipe.
type PipeMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// PipeSpec contains the spec for the Pipe.
type PipeSpec struct {
	Source PipeSource `yaml:"source"`
	Sink   PipeSink   `yaml:"sink"`
}

// PipeSource represents the source configuration.
type PipeSource struct {
	Ref        PipeRef           `yaml:"ref"`
	Properties map[string]string `yaml:"properties,omitempty"`
}

// PipeSink represents the sink configuration.
type PipeSink struct {
	Ref        PipeRef           `yaml:"ref"`
	Properties map[string]string `yaml:"properties,omitempty"`
}

// PipeRef represents a reference to a Kamelet, Broker, or endpoint.
type PipeRef struct {
	Kind       string `yaml:"kind,omitempty"`
	APIVersion string `yaml:"apiVersion,omitempty"`
	Name       string `yaml:"name,omitempty"`
	URI        string `yaml:"uri,omitempty"`
}

// Integration represents a camel-k Integration manifest.
type Integration struct {
	APIVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   IntegrationMetadata `yaml:"metadata"`
	Spec       IntegrationSpec     `yaml:"spec"`
}

// IntegrationMetadata contains metadata for the Integration.
type IntegrationMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// IntegrationSpec contains the spec for the Integration.
type IntegrationSpec struct {
	Flows   []IntegrationFlow      `yaml:"flows,omitempty"`
	Sources []IntegrationSource    `yaml:"sources,omitempty"`
	Traits  map[string]interface{} `yaml:"traits,omitempty"`
}

// IntegrationFlow represents a flow in the Integration.
type IntegrationFlow struct {
	From  IntegrationFrom          `yaml:"from"`
	Steps []map[string]interface{} `yaml:"steps,omitempty"`
	To    string                   `yaml:"to,omitempty"`
}

// IntegrationFrom represents the from clause in a flow.
type IntegrationFrom struct {
	URI string `yaml:"uri"`
}

// IntegrationSource represents a source code snippet for the Integration.
type IntegrationSource struct {
	Name     string `yaml:"name"`
	Content  string `yaml:"content"`
	Language string `yaml:"language,omitempty"`
}

// GeneratePipe generates a Pipe YAML manifest for connecting Pub/Sub to a Broker.
func GeneratePipe(cfg *Config, namespace string) (*Pipe, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if cfg.PubSubTopic == "" {
		return nil, fmt.Errorf("pubsub_topic is required")
	}

	broker, err := cfg.GetBroker()
	if err != nil {
		return nil, fmt.Errorf("get broker: %w", err)
	}

	pipeName := fmt.Sprintf("%s-pubsub-pipe", cfg.ServiceName)
	subscriptionName := fmt.Sprintf("%s-subscription", cfg.ServiceName)

	// Get GCP project ID (we'll need this for the Pub/Sub Kamelet properties)
	// For now, we'll use a placeholder that can be configured
	// The actual project will be determined at runtime from ADC

	pipe := &Pipe{
		APIVersion: "camel.apache.org/v1",
		Kind:       "Pipe",
		Metadata: PipeMetadata{
			Name:      pipeName,
			Namespace: namespace,
		},
		Spec: PipeSpec{
			Source: PipeSource{
				Ref: PipeRef{
					Kind:       "Kamelet",
					APIVersion: "camel.apache.org/v1",
					Name:       "google-pubsub-source",
				},
				Properties: map[string]string{
					"topicName":        cfg.PubSubTopic,
					"subscriptionName": subscriptionName,
				},
			},
			Sink: PipeSink{
				Properties: map[string]string{
					"cloudEventsType": fmt.Sprintf("com.stax.%s.pubsub", cfg.ServiceName),
				},
				Ref: PipeRef{
					Kind:       "Broker",
					APIVersion: "eventing.knative.dev/v1",
					Name:       broker,
				},
			},
		},
	}

	return pipe, nil
}

// ToYAML converts the Pipe to YAML.
func (p *Pipe) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

// GenerateIntegration generates an Integration YAML manifest.
func GenerateIntegration(cfg *Config, serviceName, namespace string) (*Integration, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if cfg.PubSubTopic == "" {
		return nil, fmt.Errorf("pubsub_topic is required")
	}

	integrationName := fmt.Sprintf("%s-pubsub-integration", cfg.ServiceName)

	// Generate a simple Integration that routes from pub/sub to Knative service
	integration := &Integration{
		APIVersion: "camel.apache.org/v1",
		Kind:       "Integration",
		Metadata: IntegrationMetadata{
			Name:      integrationName,
			Namespace: namespace,
		},
		Spec: IntegrationSpec{
			Flows: []IntegrationFlow{
				{
					From: IntegrationFrom{
						URI: fmt.Sprintf("google-pubsub:subscription?projectId=default&subscriptionName=%s-subscription&topicName=%s", cfg.ServiceName, cfg.PubSubTopic),
					},
					To: fmt.Sprintf("knative:endpoint/%s", serviceName),
				},
			},
			Traits: map[string]interface{}{
				"knative": map[string]interface{}{
					"enabled": true,
					"sinkBinding": map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
	}

	return integration, nil
}

// ToYAML converts the Integration to YAML.
func (i *Integration) ToYAML() ([]byte, error) {
	return yaml.Marshal(i)
}
