# {{.ProjectName}}

Stax automation project (Python, UV, stax-sdk).

## Setup

- Install [UV](https://docs.astral.sh/uv/).
- From this directory: `uv sync`

## Develop

- Run Ruff: `uv run ruff check .`
- Run tests: `uv run pytest`
- Generate Pydantic model from config: `stax generate pydantic`
- Trigger locally: `stax run --payload '{}'` (or --payload-file, --interactive)
