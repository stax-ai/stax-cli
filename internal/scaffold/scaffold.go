// Package scaffold creates new Python automation project directories from templates
// (config.yml, pyproject.toml, main.py, tests, GitHub Actions). Used by the create command.
package scaffold

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/stax-ai/stax-cli/internal/automation"
	"github.com/stax-ai/stax-cli/templates"
)

const defaultTemplate = "python"

// templateData is passed to text/template when rendering scaffold files.
type templateData struct {
	ProjectName   string
	AutomationID  string
	ProjectNamePy string // project name with dashes replaced by underscores (for pyproject name)
}

// Options configures a new project scaffold.
type Options struct {
	ProjectName  string // Directory and project name (required).
	AutomationID string // automation_id in config.yml; default is ProjectName.
	TargetDir    string // Parent directory to create project in (default: current dir).
}

// Scaffold creates a new Python automation project at TargetDir/ProjectName with config.yml,
// pyproject.toml (UV, stax-sdk, pydantic, ruff, pytest), handler, tests, and GitHub Actions.
func Scaffold(ctx context.Context, opts Options) (string, error) {
	if opts.ProjectName == "" {
		return "", fmt.Errorf("project name is required")
	}
	if opts.AutomationID == "" {
		opts.AutomationID = opts.ProjectName
	}
	root := opts.TargetDir
	if root == "" {
		root = "."
	}
	projectRoot := filepath.Join(root, opts.ProjectName)
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		return "", fmt.Errorf("create project dir: %w", err)
	}

	tplFS, ok := templates.FS[defaultTemplate]
	if !ok {
		return "", fmt.Errorf("template %q not found", defaultTemplate)
	}

	data := templateData{
		ProjectName:   opts.ProjectName,
		AutomationID:  opts.AutomationID,
		ProjectNamePy: strings.ReplaceAll(opts.ProjectName, "-", "_"),
	}

	if err := writeTemplateFS(projectRoot, defaultTemplate, tplFS, data); err != nil {
		return "", err
	}

	// Empty dirs not present in template tree
	for _, d := range []string{"src"} {
		if err := os.MkdirAll(filepath.Join(projectRoot, d), 0755); err != nil {
			return "", fmt.Errorf("create %s: %w", d, err)
		}
	}

	abs, _ := filepath.Abs(projectRoot)
	return abs, nil
}

func writeTemplateFS(projectRoot, templateName string, tplFS fs.FS, data templateData) error {
	prefix := templateName + "/" // embed.FS uses forward slash
	return fs.WalkDir(tplFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := fs.ReadFile(tplFS, path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}
		tpl, err := template.New(path).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", path, err)
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute template %s: %w", path, err)
		}
		// Embed FS paths are like "python/config.yml"; strip template prefix for project-relative path.
		relPath := path
		if strings.HasPrefix(path, prefix) {
			relPath = path[len(prefix):]
		}
		dest := filepath.Join(projectRoot, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", path, err)
		}
		if err := os.WriteFile(dest, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		return nil
	})
}

// ValidateOptions checks options before scaffolding. Returns an error if opts is nil or ProjectName is empty.
func ValidateOptions(opts *Options) error {
	if opts == nil {
		return fmt.Errorf("options is nil")
	}
	if opts.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	// Reuse automation type enum for any future field validation
	_ = automation.ValidFieldTypes
	return nil
}
