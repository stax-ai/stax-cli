## stax-cli generate

Generate code from config.yml

### Synopsis

Generate artifacts from the project config.yml. Subcommands:

  pydantic   Write a Pydantic model (AutomationInput) whose fields match config.yml.

Config is resolved from the current directory or parent dirs, or from -c/--config
when running a subcommand.

### Options

```
  -h, --help   help for generate
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli](stax-cli.md)	 - CLI for Stax automations (create, generate, run, build)
* [stax-cli generate pydantic](stax-cli_generate_pydantic.md)	 - Generate a Pydantic model from config.yml fields

