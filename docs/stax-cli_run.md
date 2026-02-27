## stax-cli run

Run the automation handler once with a CloudEvent (one-shot)

### Synopsis

Run loads config.yml (current or parent dir, or -c), builds a CloudEvent
with the given payload as event data, and invokes the project's handler via
"uv run python main.py" with the CloudEvent JSON on stdin. The handler must
be decorated with stax-sdk so it validates the CloudEvent.

You must provide exactly one of:
  -f, --payload-file   Path to a JSON file whose contents become event data
  --payload            Inline JSON object for event data
  -i, --interactive    Prompt for each field defined in config.yml

```
stax-cli run [flags]
```

### Examples

```
  stax run --payload '{}'
  stax run -f data.json
  stax run -c ./project/config.yml --payload '{"key":"value"}'
```

### Options

```
  -c, --config string         path to config.yml (default: search in current and parent dirs)
  -h, --help                  help for run
  -i, --interactive           prompt for each field from config
      --payload string        inline JSON object for event data
  -f, --payload-file string   path to JSON file for event data
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli](stax-cli.md)	 - CLI for Stax automations (create, generate, run, build)

