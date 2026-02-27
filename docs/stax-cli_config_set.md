## stax-cli config set

Set a config value

### Synopsis

Writes KEY=VALUE to the global config file. Creates the file if it does not exist.

```
stax-cli config set KEY VALUE [flags]
```

### Examples

```
  stax config set api_url https://api.example.com
  stax config set profile default
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli config](stax-cli_config.md)	 - Read or write stax CLI configuration

