# Repository Guidelines

[中文](AGENTS.zh-CN.md) | English

## Project Structure & Module Organization

`cmd/godesk/main.go` is the CLI entrypoint. Cobra command wiring and the dependency-free TUI live in `internal/cli`. Shared behavior is split by responsibility: `internal/config` handles global config, scan roots, project index, and `.godesk.yaml`; `internal/project` handles Go module scanning and file discovery; `internal/envfile` and `internal/compose` parse project files; `internal/docker`, `internal/logtail`, `internal/ports`, and `internal/runner` wrap runtime integrations. User docs live in `README.md` and `README.zh-CN.md`; developer docs live in `docs/`.

## Build, Test, and Development Commands

Use these commands from the repository root:

```bash
go run ./cmd/godesk --help        # run the CLI from source
go run ./cmd/godesk <cmd> --help  # inspect one command
go build -o godesk ./cmd/godesk   # build a local binary
go run ./cmd/godesk list          # smoke-check config/index loading
```

For Docker-related changes, verify the host tools directly:

```bash
docker version
docker compose version
lsof -v
```

## Coding Style & Naming Conventions

Format Go files with `gofmt`. Keep command constructors named `new<Name>Command` and register them in `internal/cli/root.go`. Project commands use `godesk <command> <project>` and validate with `requireProjectName`. Global config commands use explicit subcommands such as `godesk roots add <path>`. Validation commands should report `ok`, `warn`, and `fail` consistently. Put reusable logic in focused internal packages rather than inside command handlers.

## Testing Guidelines

Use command-level smoke checks for current verification. For CLI changes, run `go run ./cmd/godesk --help` and the relevant `go run ./cmd/godesk <command> --help`. For commands that write files or user config, verify the printed path and resolved project values. For Docker and port behavior, confirm Docker or `lsof` behavior before debugging godesk code.

## Commit & Pull Request Guidelines

Existing history uses concise conventional-style subjects, such as `init: scaffold godesk project` and `refactor: add init commands, remove test cmd, improve project discovery`. Prefer `type: summary` with a short imperative summary. Pull requests should describe the user-facing command behavior, list touched packages, include smoke-check commands, and note any config file changes.

## Configuration Tips

Project config belongs in `.godesk.yaml` at the Go module root. Store project file paths relative to that root, for example `compose_file: docker/docker-compose.yml` and `log_files: ["./logs/app.log"]`. Store runnable commands as shell commands, for example `lint_cmd: "go vet ./... && golangci-lint run"`.
