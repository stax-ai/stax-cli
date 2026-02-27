## stax-cli build

Build a container image from the project using buildpacks

### Synopsis

Build produces an OCI container image from the current automation project
using Cloud Native Buildpacks (pack). Requires Docker (or a compatible daemon) to be running.

The project directory must contain at least main.py and pyproject.toml (and typically
config.yml). Optionally, .staxignore lists paths to exclude from the build context.
A Procfile defines the web process (e.g. "web: python main.py").

```
stax-cli build [flags]
```

### Examples

```
  stax build --image myreg.io/my-automation:latest
  stax build --path ./my-project --registry myreg.io
  stax build --image localhost:5000/app:v1 --builder ghcr.io/knative/builder-jammy-base:v2 --verbose
```

### Options

```
      --builder string    builder image (default: ghcr.io/knative/builder-jammy-base:v2)
  -h, --help              help for build
      --image string      full OCI image name to build (e.g. myreg.io/myapp:latest); required unless --registry is set with a project that has config.yml
      --path string       project root (default: current directory) (default ".")
      --registry string   registry for deriving image name when --image is not set (image = registry/automation_id:latest from config.yml)
      --trust-builder     trust any builder image (default: only trusted prefixes)
      --verbose           stream build logs to stderr
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli](stax-cli.md)	 - CLI for Stax automations (create, generate, run, build)

