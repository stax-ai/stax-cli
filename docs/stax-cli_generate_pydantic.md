## stax-cli generate pydantic

Generate a Pydantic model from config.yml fields

### Synopsis

Read config.yml (from cwd/parent dirs or -c/--config), validate it, and write
a Python file containing a Pydantic BaseModel (default class name: AutomationInput)
whose fields match the config's fields. Types are mapped (e.g. integer -> int,
datetime -> datetime). Optional fields become Optional[...] = None.

Output path is relative to the project root unless an absolute path is given.

```
stax-cli generate pydantic [flags]
```

### Examples

```
  stax generate pydantic
  stax generate pydantic -c ./project/config.yml --output src/models.py
```

### Options

```
  -c, --config string   path to config.yml (default: search config.yml in current and parent dirs)
  -h, --help            help for pydantic
      --output string   output Python file path (default "src/generated_models.py")
```

### Options inherited from parent commands

```
      --config-file string   path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)
```

### SEE ALSO

* [stax-cli generate](stax-cli_generate.md)	 - Generate code from config.yml

