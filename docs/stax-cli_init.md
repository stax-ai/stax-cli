## stax-cli init

Initialize an existing repo with UV, linting, GitHub Actions, and config.yml

### Synopsis

Initialize the current directory (or --path) with Stax automation layout:

  - config.yml       Minimal automation config (automation_id, handler, empty fields)
  - pyproject.toml   UV project with stax-sdk, pydantic, ruff, pytest
  - main.py          Thin entrypoint importing handler from package (if missing)
  - src/<package>/   Package with stub handler (if missing)
  - .github/workflows/ci.yml   CI (UV, Ruff, pytest)
  - .python-version, .staxignore, Procfile

Does not overwrite existing files unless --force (applies to config.yml and CI only).

```
stax-cli init [flags]
```

### Examples

```
  stax init
  stax init --path ./my-repo --automation-id my-id
  stax init --force
```

### Options

```
      --automation-id string   automation_id in config.yml (default: directory name)
      --force                  overwrite existing config.yml and .github/workflows/ci.yml
  -h, --help                   help for init
      --path string            directory to initialize (default: current directory)
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli](stax-cli.md)	 - CLI for Stax automations (create, init, generate, run, build)

