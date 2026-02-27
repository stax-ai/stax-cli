"""Thin entry point: imports handler from package for Procfile/run."""
import json
import sys

from {{.ProjectNamePy}} import handler

if __name__ == "__main__":
    payload = json.load(sys.stdin)
    result = handler(payload)
    print(json.dumps(result))
