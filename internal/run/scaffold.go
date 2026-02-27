// Package run: WriteRunScaffold writes transient scaffolding into a run directory
// so the Knative func-python wrapper can be started without modifying the user's main.py.
package run

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	runScaffoldDir     = ".stax"
	runScaffoldRunsDir = "runs"
)

// serviceMainPy is the transient service/main.py that imports the user's handler
// and wraps it for func_python.cloudevent.serve().
const serviceMainPy = `"""
Transient wrapper: loads the user's handler from STAX_HANDLER (e.g. main:handler),
adapts it to func-python's async handle(scope, receive, send), and runs the ASGI server.
"""
import asyncio
import importlib
import logging
import os
import sys

# Project root (symlink f) must be on path before importing user handler.
_run_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_f = os.path.join(_run_dir, "f")
if _f not in sys.path:
    sys.path.insert(0, _f)

logging.basicConfig(level=logging.INFO)

_handler_spec = os.environ.get("STAX_HANDLER", "main:handler")
_mod_name, _, _fn_name = _handler_spec.partition(":")
if not _fn_name:
    _fn_name = "handler"
_mod = importlib.import_module(_mod_name)
_user_handler = getattr(_mod, _fn_name)


def _event_to_dict(event):
    """Convert cloudevents SDK CloudEvent to dict for stax handler."""
    data = getattr(event, "data", None)
    if data is None and hasattr(event, "get_data"):
        try:
            data = event.get_data()
        except Exception:
            pass
    if data is None:
        data = {}
    out = {"specversion": getattr(event, "specversion", "1.0"), "id": getattr(event, "id", ""),
           "source": getattr(event, "source", ""), "type": getattr(event, "type", ""), "data": data}
    return out


async def handle(scope, receive, send):
    from func_python.cloudevent import decode_event
    from cloudevents.http import CloudEvent
    event = scope.get("event")
    if event is None:
        scope["event"] = await decode_event(scope, receive)
        event = scope["event"]
    event_dict = _event_to_dict(event)
    result = await asyncio.to_thread(_user_handler, event_dict)
    if result is None:
        result = {}
    if not isinstance(result, dict):
        result = {"result": result}
    response = CloudEvent(
        {"type": "stax.cli/response", "source": "stax-cli/run"},
        result,
    )
    await send(response)


if __name__ == "__main__":
    from func_python.cloudevent import serve
    logging.info("Stax run middleware invoking user function")
    serve(handle)
`

// runScaffoldPyproject is the transient pyproject.toml for the run environment.
const runScaffoldPyproject = `[project]
name = "stax-run-service"
version = "0.0.1"
requires-python = ">=3.9"
dependencies = [
    "func-python>=0.7.0",
    "cloudevents",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
`

// WriteRunScaffold writes the transient scaffolding into runDir and creates
// a symlink "f" from runDir to projectRoot. handlerSpec is the module:function
// (e.g. "main:handler"); it is passed via STAX_HANDLER when starting the process.
func WriteRunScaffold(runDir, projectRoot, handlerSpec string) error {
	if runDir == "" || projectRoot == "" {
		return fmt.Errorf("runDir and projectRoot are required")
	}
	serviceDir := filepath.Join(runDir, "service")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("create service dir: %w", err)
	}
	mainPy := filepath.Join(serviceDir, "main.py")
	if err := os.WriteFile(mainPy, []byte(serviceMainPy), 0644); err != nil {
		return fmt.Errorf("write service/main.py: %w", err)
	}
	pyprojectPath := filepath.Join(runDir, "pyproject.toml")
	if err := os.WriteFile(pyprojectPath, []byte(runScaffoldPyproject), 0644); err != nil {
		return fmt.Errorf("write pyproject.toml: %w", err)
	}
	linkPath := filepath.Join(runDir, "f")
	_ = os.Remove(linkPath)
	rel, err := filepath.Rel(runDir, projectRoot)
	if err != nil {
		return fmt.Errorf("relative path from runDir to projectRoot: %w", err)
	}
	if err := os.Symlink(rel, linkPath); err != nil {
		return fmt.Errorf("symlink f -> %s: %w", rel, err)
	}
	_ = handlerSpec // passed via STAX_HANDLER when running
	return nil
}

// RunDir returns the run directory for the given project root and port.
// It does not create the directory.
func RunDir(projectRoot, port string) string {
	return filepath.Join(projectRoot, runScaffoldDir, runScaffoldRunsDir, port)
}
