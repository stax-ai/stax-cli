package templates

import (
	"embed"
	"io/fs"
)

//go:embed all:python
var pythonFS embed.FS

// FS maps CLI template name (e.g. "python") to the embedded template filesystem for that set.
// Use when adding a new template: add a dir templates/<name>/, an unexported var with //go:embed all:<name>, and an entry here.
var FS = map[string]fs.FS{
	"python": pythonFS,
}
