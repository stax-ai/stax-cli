// Package automation parses and validates project-level config.yml (automation_id, handler, fields).
// It is used by create, generate, and run. Field types match JSON Schema and map to Pydantic when generating.
package automation

// FieldType constants are the allowed type strings for automation config fields.
// Matches JSON Schema–style types and maps to Pydantic when generating.
const (
	FieldTypeString   = "string"
	FieldTypeInteger  = "integer"
	FieldTypeNumber   = "number"
	FieldTypeBoolean  = "boolean"
	FieldTypeArray    = "array"
	FieldTypeObject   = "object"
	FieldTypeDatetime = "datetime"
	FieldTypeUser = "User"
	FieldTypeDropdown = "Dropdown"

)

// ValidFieldTypes is the list of allowed type strings for validation.
var ValidFieldTypes = []string{
	FieldTypeString, FieldTypeInteger, FieldTypeNumber, FieldTypeBoolean,
	FieldTypeArray, FieldTypeObject, FieldTypeDatetime, FieldTypeUser, FieldTypeDropdown,
} 

// Config is the project-level automation config (config.yml).
// JSON Schema–style YAML: automation_id, handler, fields.
type Config struct {
	AutomationID string  `yaml:"automation_id"`
	Handler      string  `yaml:"handler"`
	Fields       []Field `yaml:"fields"`
}

// Field defines one input field for the automation.
type Field struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description,omitempty"`
	Default     any    `yaml:"default,omitempty"`
	Linkable    bool   `yaml:"linkable,omitempty"`
	Options     []string `yaml:"options,omitempty"`
}
