package run

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stax-ai/stax-cli/internal/automation"
)

// PayloadFromInteractive prompts for each field in cfg on stderr/stdin and returns
// a map of field names to parsed values. Type-aware: integer, number, boolean, etc.
// For array/object, expects JSON on a single line.
func PayloadFromInteractive(cfg *automation.Config) (map[string]any, error) {
	if cfg == nil || len(cfg.Fields) == 0 {
		return make(map[string]any), nil
	}
	scanner := bufio.NewScanner(os.Stdin)
	out := make(map[string]any)
	for _, f := range cfg.Fields {
		prompt := f.Name
		if f.Description != "" {
			prompt = fmt.Sprintf("%s (%s)", f.Name, f.Description)
		}
		if !f.Required {
			prompt += " [optional]"
		}
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		if !scanner.Scan() {
			return nil, scanner.Err()
		}
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			if f.Required {
				return nil, fmt.Errorf("field %s is required", f.Name)
			}
			if f.Default != nil {
				out[f.Name] = f.Default
			}
			continue
		}
		val, err := parseFieldValue(raw, f.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", f.Name, err)
		}
		out[f.Name] = val
	}
	return out, nil
}

func parseFieldValue(raw, typ string) (any, error) {
	switch typ {
	case automation.FieldTypeString, automation.FieldTypeDatetime:
		return raw, nil
	case automation.FieldTypeInteger:
		n, err := strconv.ParseInt(raw, 10, 64)
		return n, err
	case automation.FieldTypeNumber:
		f, err := strconv.ParseFloat(raw, 64)
		return f, err
	case automation.FieldTypeBoolean:
		lower := strings.ToLower(raw)
		if lower == "true" || lower == "1" || lower == "yes" {
			return true, nil
		}
		if lower == "false" || lower == "0" || lower == "no" || lower == "" {
			return false, nil
		}
		return nil, fmt.Errorf("invalid boolean: %q", raw)
	case automation.FieldTypeArray, automation.FieldTypeObject:
		// For v1, treat as JSON string
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return nil, fmt.Errorf("invalid JSON for %s: %w", typ, err)
		}
		return v, nil
	default:
		return raw, nil
	}
}
