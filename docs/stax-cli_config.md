## stax-cli config

Read or write stax CLI configuration

### Synopsis

Manage global stax-cli configuration stored in .stax.yml (or .stax.yaml).
Config is searched in: user config dir, home, then current directory. Use the
global --config-file flag to point to a specific file.

Subcommands:
  get   Print a config key or the entire config (YAML)
  set   Set a config key to a value

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli](stax-cli.md)	 - CLI for Stax automations (create, generate, run, build)
* [stax-cli config get](stax-cli_config_get.md)	 - Get a config value or the entire config
* [stax-cli config set](stax-cli_config_set.md)	 - Set a config value

