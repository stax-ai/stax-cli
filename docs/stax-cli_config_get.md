## stax-cli config get

Get a config value or the entire config

### Synopsis

With no KEY, prints the full config as YAML. With KEY, prints that key's value.

```
stax-cli config get [KEY] [flags]
```

### Examples

```
  stax config get
  stax config get api_url
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli config](stax-cli_config.md)	 - Read or write stax CLI configuration

