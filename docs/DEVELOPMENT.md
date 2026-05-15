# godesk Development Guide

[中文](DEVELOPMENT.zh-CN.md) | English

## Purpose

godesk is a local Go backend workspace manager. It helps developers scan local Go projects, resolve project config, start dependency services, inspect ports, and run configured project tooling.

The current product surface is CLI. TUI can build on top of the same internal packages later.

## Environment

Required tools:

```text
Go 1.25+
```

Runtime integrations:

```text
Docker CLI and Docker Compose for `godesk up`
lsof for `godesk ports`
```

Useful local checks:

```bash
go run ./cmd/godesk --help
go run ./cmd/godesk list
docker version
docker compose version
lsof -v
```

## Repository Layout

```text
cmd/godesk/main.go       CLI entrypoint
internal/cli             Cobra command wiring and command handlers
internal/config          Global config, project index, and .godesk.yaml handling
internal/project         Go project scanning and file discovery
internal/envfile         .env parser
internal/compose         Docker Compose parser
internal/docker          Docker Compose process integration
internal/ports           Port candidate extraction and lsof checks
internal/runner          Custom command runner
docs                     Development documentation
```

Keep shared behavior in internal packages so CLI and future TUI code can use the same discovery, parsing, Docker, port, and runner logic.

## Command Model

Commands are grouped into project commands and global commands.

Project commands use one standard form:

```bash
godesk <command> <project>
```

Current project commands:

```bash
godesk init <project>
godesk inspect <project>
godesk up <project>
godesk ports <project>
godesk health <project>
godesk lint <project>
```

Global commands operate on the workspace index or scan roots:

```bash
godesk list
godesk scan [root...]
```

Commands for indexed projects should accept the project name in the same positional location as the existing project commands.

Commands for direct local workflows should use a separate explicit command name. Example:

```bash
godesk init-local
```

## Config Files

godesk stores its global files under `os.UserConfigDir()/godesk`.

On macOS this usually resolves to:

```text
~/Library/Application Support/godesk
```

Current global files:

```text
config.yaml    scan roots
projects.yaml  scanned project index
```

Each Go project can define `.godesk.yaml` in the module root:

```yaml
name: my-project
env_file: .env
compose_file: docker/docker-compose.yml
lint_cmd: golangci-lint run
up_cmd: make up
health_urls:
  - http://localhost:8080/health
```

`godesk init <project>` generates `.godesk.yaml` for an indexed project. The project name comes from the index record, and discovered paths are written relative to the project root.

`godesk init-local` generates `.godesk.yaml` for the current Go project by walking upward to `go.mod`.

Project config overrides scan results when the project is scanned again.

## Scan And Discovery

Scanning finds Go projects by locating directories that contain `go.mod`.

Project discovery resolves these files together:

```text
.env
docker-compose.yml
docker-compose.yaml
compose.yaml
```

Discovery searches from the Go module directory upward to the scan root, then searches downward from the project root with a bounded depth. This supports layouts like:

```text
project/
  go.mod
  docker/
    docker-compose.yml
```

When multiple downward matches exist, prefer the shallowest match, then the lexicographically first path.

## Runtime Behavior

`up` starts dependency services for a project. It uses `up_cmd` when configured; otherwise it uses Docker Compose with the resolved compose file.

`ports` collects candidate ports from `.env` entries with port-like keys and from Docker Compose published ports, then reports local listener status.

`health` checks `health_urls` from `.godesk.yaml` and reports status code, latency, and request errors.

`lint` runs `lint_cmd` from `.godesk.yaml` when configured.

`inspect` prints resolved project config, parsed env entries, and compose services.

## Adding A Project Command

Use this flow for a command that operates on one indexed project:

1. Add a new file under `internal/cli`, for example `health.go`.
2. Define `newHealthCommand(app *appContext) *cobra.Command`.
3. Set `Use` to `health <project>`.
4. Use `requireProjectName(cmd, args)` for argument validation.
5. Resolve the project with `app.store.FindProject(name)`.
6. Put reusable behavior in a focused internal package when the logic can be shared.
7. Register the command in `internal/cli/root.go`.
8. Document the command in `README.md`, `README.zh-CN.md`, and this guide.
9. Verify with help output and one command-level smoke check.

Project command skeleton:

```go
func newHealthCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "health <project>",
		Short: "Show project health status",
		Args: func(cmd *cobra.Command, args []string) error {
			_, err := requireProjectName(cmd, args)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := requireProjectName(cmd, args)
			p, err := app.store.FindProject(name)
			if err != nil {
				return err
			}
			_ = p
			return nil
		},
	}
}
```

## Adding A Global Command

Use a global command for behavior that operates on the whole index, config, or scan roots.

1. Add the command file under `internal/cli`.
2. Set `Args` to `cobra.NoArgs` or define explicit positional args.
3. Load config or index through `app.store`.
4. Keep output tabular when the command lists resources.
5. Register the command in `internal/cli/root.go`.
6. Update user and developer docs.

Examples of global commands:

```bash
godesk list
godesk scan [root...]
```

## Adding A Project Config Field

Use this flow when a new field belongs in `.godesk.yaml`:

1. Add the field to `project.Project` in `internal/project/project.go`.
2. Add the field to `config.ProjectOverride` in `internal/config/config.go`.
3. Apply it in `config.ApplyOverride`.
4. Write it in `internal/cli/init.go` when generating `.godesk.yaml`.
5. Use the field from the relevant command or internal package.
6. Add the field to README examples and both development guides.
7. Verify `godesk init <project>` or `godesk init-local` output.

Use relative paths for project files stored in `.godesk.yaml`.

## Adding Discovery Rules

Keep discovery logic in `internal/project`.

When adding a new project file type, update discovery so related files are found in the same walk where possible. This keeps scan behavior predictable and efficient.

Preferred matching order for downward discovery:

```text
1. shallowest path
2. lexicographically first path
```

Use bounded downward search for project-local files that may live in subdirectories.

## Output Style

Use compact terminal output.

For multi-row output, use `text/tabwriter` with a header row.

For project details, print stable labels:

```text
name:
path:
env:
compose:
lint:
up:
health:
```

Use `-` for missing optional values.

## Common Local Workflows

Scan and inspect a workspace:

```bash
go run ./cmd/godesk scan /path/to/workspace
go run ./cmd/godesk list
go run ./cmd/godesk inspect <project>
```

Initialize project config:

```bash
go run ./cmd/godesk init <project>
go run ./cmd/godesk init-local
```

Check runtime state:

```bash
go run ./cmd/godesk ports <project>
go run ./cmd/godesk up <project>
go run ./cmd/godesk health <project>
```

Run configured tooling:

```bash
go run ./cmd/godesk lint <project>
```

## Debugging

For scan issues, compare:

```bash
go run ./cmd/godesk scan /path/to/workspace
go run ./cmd/godesk list
go run ./cmd/godesk inspect <project>
```

For discovery issues, inspect the project layout and `.godesk.yaml`:

```bash
find /path/to/project -maxdepth 4 -name go.mod -o -name .env -o -name 'docker-compose.yml' -o -name 'docker-compose.yaml' -o -name 'compose.yaml'
```

For Docker issues, verify Docker directly:

```bash
docker version
docker compose version
docker compose -f /path/to/docker-compose.yml config
```

For port issues, verify the port directly:

```bash
lsof -nP -iTCP:<port> -sTCP:LISTEN
```

## Verification

For narrow CLI changes, verify with command-level smoke checks:

```bash
go run ./cmd/godesk --help
go run ./cmd/godesk <command> --help
go run ./cmd/godesk list
```

For commands that write user config or project files, verify the generated output path and the displayed resolved project/env/compose values.

For Docker behavior, verify Docker CLI availability and Docker daemon access before debugging godesk-specific behavior.

## Current Scope And Next Extensions

Current implemented scope:

```text
scan
list
init
init-local
inspect
up
ports
health
lint
```

Natural next extensions:

```text
multi-log tail command
TUI project dashboard
Docker service status view
global config editing command
```
