## stax-cli

CLI for Stax automations (create, generate, run, build)

### Synopsis

stax-cli manages Stax automation projects: scaffold Python projects with
config.yml, generate Pydantic models from config, run handlers locally
with CloudEvents, and build container images with buildpacks.

Commands:
  config   Read or write global CLI config (.stax.yml)
  create   Scaffold a new Python automation project (UV, stax-sdk, Ruff, pytest)
  generate Generate code from config.yml (e.g. Pydantic models)
  run      Invoke the automation handler once with a CloudEvent (one-shot)
  build    Build a container image from the project using Cloud Native Buildpacks

Global flags (e.g. --config-file) apply to config get/set. Project commands
(create, generate, run, build) use config.yml in the project directory.

### Options

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
  -h, --help                 help for stax-cli
```

### SEE ALSO

* [stax-cli build](stax-cli_build.md)	 - Build a container image from the project using buildpacks
* [stax-cli config](stax-cli_config.md)	 - Read or write stax CLI configuration
* [stax-cli create](stax-cli_create.md)	 - Scaffold a new Python automation project
* [stax-cli generate](stax-cli_generate.md)	 - Generate code from config.yml
* [stax-cli run](stax-cli_run.md)	 - Run the automation handler once with a CloudEvent (one-shot)

