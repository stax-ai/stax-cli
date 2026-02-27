package automation

import (
	"go.yaml.in/yaml/v3"
)

// unmarshalYAML decodes YAML into v. Wrapped for consistency and testing.
func unmarshalYAML(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
