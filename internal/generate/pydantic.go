// Package generate produces code from automation config (e.g. Pydantic models from config.yml fields).
// Used by the generate command.
package generate

import (
	"fmt"
	"os"
	"strings"

	"github.com/stax-ai/stax-cli/internal/automation"
)

const defaultModelName = "AutomationInput"

// GeneratePydantic writes a Python file containing a Pydantic BaseModel whose fields
// match the given automation config. Parent directories of outputPath are created if needed.
// Returns an error if cfg is nil or writing fails.
func GeneratePydantic(cfg *automation.Config, outputPath string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	content := pydanticContent(cfg)
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// pydanticContent returns the full Python source for the Pydantic model.
func pydanticContent(cfg *automation.Config) string {
	var b strings.Builder
	b.WriteString(`"""Generated Pydantic model from config.yml. Do not edit by hand."""
from typing import Any, Optional

`)
	// Add datetime import only if needed
	needDatetime := false
	for _, f := range cfg.Fields {
		if f.Type == automation.FieldTypeDatetime {
			needDatetime = true
			break
		}
	}
	if needDatetime {
		b.WriteString("from datetime import datetime\n\n")
	}
	b.WriteString("from pydantic import BaseModel\n\n\n")
	b.WriteString("class ")
	b.WriteString(defaultModelName)
	b.WriteString("(BaseModel):\n")
	for _, f := range cfg.Fields {
		b.WriteString("    ")
		b.WriteString(fieldLine(f))
		b.WriteString("\n")
	}
	return b.String()
}

func fieldLine(f automation.Field) string {
	pyType := pydanticType(f.Type)
	if !f.Required {
		pyType = "Optional[" + pyType + "] = None"
	}
	return f.Name + ": " + pyType
}

func pydanticType(t string) string {
	switch t {
	case automation.FieldTypeString:
		return "str"
	case automation.FieldTypeInteger:
		return "int"
	case automation.FieldTypeNumber:
		return "float"
	case automation.FieldTypeBoolean:
		return "bool"
	case automation.FieldTypeArray:
		return "list[Any]"
	case automation.FieldTypeObject:
		return "dict[str, Any]"
	case automation.FieldTypeDatetime:
		return "datetime"
	default:
		return "Any"
	}
}
