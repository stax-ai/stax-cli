// Package initrepo initializes an existing directory with Stax automation layout
// (config.yml, pyproject.toml, main.py, CI, etc.). Used by the init command.
package initrepo

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/stax-ai/stax-cli/templates"
)

const templateName = "python"

// templateData is passed to text/template when rendering files.
type templateData struct {
	ProjectName   string
	AutomationID  string
	ProjectNamePy string
}

// Options configures init behavior.
type Options struct {
	ProjectRoot  string // Target directory (required).
	AutomationID string // automation_id in config.yml; default is directory base name.
	Force        bool   // Overwrite existing config.yml and .github/workflows/ci.yml.
}

// Init writes missing (or force-overwritable) Stax files into an existing directory.
// Idempotent: does not overwrite existing files unless Force is true for config/CI.
func Init(ctx context.Context, opts Options) (string, error) {
	if opts.ProjectRoot == "" {
		return "", fmt.Errorf("project root is required")
	}
	absRoot, err := filepath.Abs(opts.ProjectRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	if err := ensureDir(absRoot); err != nil {
		return "", err
	}

	baseName := filepath.Base(absRoot)
	if baseName == "." || baseName == "/" {
		var err error
		absRoot, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("project root: %w", err)
		}
		baseName = filepath.Base(absRoot)
	}

	automationID := opts.AutomationID
	if automationID == "" {
		automationID = baseName
	}
	projectNamePy := strings.ReplaceAll(baseName, "-", "_")

	data := templateData{
		ProjectName:   baseName,
		AutomationID:  automationID,
		ProjectNamePy: projectNamePy,
	}

	tplFS, ok := templates.FS[templateName]
	if !ok {
		return "", fmt.Errorf("template %q not found", templateName)
	}

	// Files that may be overwritten with --force (config and CI only).
	forceOverwritable := map[string]bool{
		"config.yml":              true,
		".github/workflows/ci.yml": true,
	}

	// Files we never overwrite (user code / project file).
	skipPaths := map[string]bool{
		"README.md": true, // user may have their own
	}

	configExists := configExistsIn(absRoot)

	prefix := templateName + "/" // embed FS paths are like "python/config.yml"
	if err := fs.WalkDir(tplFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath := path
		if strings.HasPrefix(path, prefix) {
			relPath = path[len(prefix):]
		}
		relPath = strings.ReplaceAll(relPath, "_pkg_", data.ProjectNamePy)
		dest := filepath.Join(absRoot, filepath.FromSlash(relPath))

		if skipPaths[relPath] {
			return nil
		}

		// Skip writing config.yml if either config.yml or config.yaml exists, unless --force.
		if relPath == "config.yml" && configExists && !opts.Force {
			return nil
		}

		overwrite := opts.Force && forceOverwritable[relPath]
		if _, err := os.Stat(dest); err == nil && !overwrite {
			return nil // skip existing file
		}

		var buf bytes.Buffer
		if relPath == "config.yml" {
			buf.WriteString(initConfigYAML(data))
		} else {
			content, err := fs.ReadFile(tplFS, path)
			if err != nil {
				return fmt.Errorf("read template %s: %w", path, err)
			}
			tpl, err := template.New(path).Parse(string(content))
			if err != nil {
				return fmt.Errorf("parse template %s: %w", path, err)
			}
			if err := tpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("execute template %s: %w", path, err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", relPath, err)
		}
		if err := os.WriteFile(dest, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("write %s: %w", relPath, err)
		}
		return nil
	}); err != nil {
		return "", err
	}

	return absRoot, nil
}

func configExistsIn(dir string) bool {
	for _, name := range []string{"config.yml", "config.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

// initConfigYAML returns minimal config.yml content (automation_id, handler, empty fields).
func initConfigYAML(data templateData) string {
	return fmt.Sprintf("automation_id: %s\nhandler: main:handler\nfields: []\n", data.AutomationID)
}

func ensureDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project root does not exist: %s", dir)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("project root is not a directory: %s", dir)
	}
	return nil
}
