package automation

import (
	"errors"
	"fmt"
	"strings"
)

// Validate checks the config for logical errors and returns all validation errors.
func Validate(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	var errs []string
	if cfg.AutomationID == "" {
		errs = append(errs, "automation_id is required")
	}
	if cfg.Handler == "" {
		errs = append(errs, "handler is required")
	} else if !isValidHandler(cfg.Handler) {
		errs = append(errs, "handler must be in format 'module:function'")
	}
	for i, f := range cfg.Fields {
		if f.Name == "" {
			errs = append(errs, fmt.Sprintf("fields[%d].name is required", i))
		}
		if f.Type == "" {
			errs = append(errs, fmt.Sprintf("fields[%d].type is required", i))
		} else if !isValidFieldType(f.Type) {
			errs = append(errs, fmt.Sprintf("fields[%d].type %q is not valid; must be one of: %s",
				i, f.Type, strings.Join(ValidFieldTypes, ", ")))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

func isValidHandler(h string) bool {
	parts := strings.SplitN(h, ":", 2)
	return len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != ""
}

func isValidFieldType(t string) bool {
	for _, v := range ValidFieldTypes {
		if t == v {
			return true
		}
	}
	return false
}
