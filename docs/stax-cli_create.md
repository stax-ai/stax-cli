## stax-cli create

Scaffold a new Python automation project

### Synopsis

Create a new directory named project-name with a full Python automation layout:

  - config.yml       Stub automation config (automation_id, handler, fields)
  - pyproject.toml   UV project with pydantic, stax-sdk, ruff, pytest
  - main.py          Handler decorated with stax-sdk (main:handler)
  - tests/           Pytest tests
  - .github/workflows/ci.yml   CI (UV, Ruff, pytest)

Use --path to create the project in a specific parent directory; use
--automation-id to set the automation_id in config.yml (default: project name).

```
stax-cli create [project-name] [flags]
```

### Examples

```
  stax create my-automation
  stax create my-automation --path ./repos --automation-id my-id
```

### Options

```
      --automation-id string   automation_id in config.yml (default: project name)
  -h, --help                   help for create
      --path string            parent directory to create project in (default: current directory)
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli](stax-cli.md)	 - CLI for Stax automations (create, init, generate, run, build)

