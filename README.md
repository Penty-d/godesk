# godesk

[中文](README.zh-CN.md) | English

godesk is a CLI workspace manager for local Go backend projects. It scans local Go modules, resolves project config, reads `.env` and Docker Compose files, starts dependency services, checks port occupancy, tails logs, checks health URLs, and runs configured lint commands.

## Features

- Scan local directories for Go projects by `go.mod`
- Resolve `.env` and Docker Compose files near a project
- Generate per-project `.godesk.yaml`
- Manage default scan roots
- List indexed projects
- Open an interactive terminal workspace
- Inspect resolved project config, env entries, and compose services
- Check project configuration and local tool availability
- Start dependency services with Docker Compose or a custom command
- Show local port occupancy from env and compose config
- Tail configured log files and Docker Compose service logs
- Check configured health URLs
- Run a configured lint command

## Install

Run from source:

```bash
go run ./cmd/godesk --help
```

Build a local binary:

```bash
go build -o godesk ./cmd/godesk
```

Then run:

```bash
./godesk --help
```

## Basic Workflow

Scan a directory that contains Go projects:

```bash
godesk scan /Users/penty/Desktop/Projects/fuu
```

Save a default scan root:

```bash
godesk roots add /Users/penty/Desktop/Projects/fuu
godesk scan
```

List indexed projects:

```bash
godesk list
```

Open the interactive workspace:

```bash
godesk tui
```

Generate project config:

```bash
godesk init fzuhelper-server
```

Inspect the resolved project:

```bash
godesk inspect fzuhelper-server
```

Check project setup:

```bash
godesk doctor fzuhelper-server
```

Check ports:

```bash
godesk ports fzuhelper-server
```

Check health:

```bash
godesk health fzuhelper-server
```

Tail logs:

```bash
godesk logs fzuhelper-server
```

Start dependency services:

```bash
godesk up fzuhelper-server
```

Run lint:

```bash
godesk lint fzuhelper-server
```

## Command Model

Project commands use the same shape:

```bash
godesk <command> <project>
```

Current project commands:

```bash
godesk init <project>
godesk inspect <project>
godesk doctor <project>
godesk up <project>
godesk ports <project>
godesk health <project>
godesk logs <project> [service...]
godesk lint <project>
```

Global commands:

```bash
godesk roots add <path>
godesk roots list
godesk roots remove <path>
godesk scan [root...]
godesk list
godesk tui
```

For a project that has not been scanned yet, initialize the current Go module directly:

```bash
godesk init-local
```

## Project Config

Each project can define a `.godesk.yaml` in the Go module root:

```yaml
name: fzuhelper-server
env_file: .env
compose_file: docker/docker-compose.yml
lint_cmd: "go vet ./... && golangci-lint run"
up_cmd: "make up ENV=local"
health_urls:
  - http://localhost:8080/health
log_files:
  - ./logs/app.log
  - ./logs/worker.log
```

`godesk init <project>` creates this file for an indexed project. `godesk init-local` creates it for the current Go module.

When scanning, `.godesk.yaml` overrides the discovered project values.

## Discovery

godesk discovers projects by walking scan roots and finding directories that contain `go.mod`.

For each project, it resolves these files:

```text
.env
docker-compose.yml
docker-compose.yaml
compose.yaml
```

Discovery searches upward from the Go module directory to the scan root, then searches downward from the project root with a bounded depth. This supports layouts such as:

```text
project/
  go.mod
  docker/
    docker-compose.yml
```

## Commands

### `scan`

Scan roots and save the project index:

```bash
godesk scan /path/to/workspace
```

The older flag form is also supported:

```bash
godesk scan --root /path/to/workspace
```

### `list`

Show indexed projects:

```bash
godesk list
```

### `roots`

Manage default scan roots:

```bash
godesk roots add /path/to/workspace
godesk roots list
godesk roots remove /path/to/workspace
```

After adding roots, scan can run without arguments:

```bash
godesk scan
```

### `tui`

Open the interactive project workspace:

```bash
godesk tui
```

Keyboard controls:

```text
j/k or arrows  move project selection
r              reload project index
i              run inspect for the selected project
p              run ports for the selected project
h              run health for the selected project
u              run up for the selected project
l              run logs for the selected project
q              quit
```

### `init`

Create `.godesk.yaml` for an indexed project:

```bash
godesk init <project>
```

Overwrite an existing config:

```bash
godesk init --force <project>
```

### `init-local`

Create `.godesk.yaml` for the current Go module:

```bash
godesk init-local
```

### `inspect`

Print resolved project details:

```bash
godesk inspect <project>
```

### `doctor`

Check project configuration and local tool availability:

```bash
godesk doctor <project>
```

Doctor checks the project root, `go.mod`, configured env file, compose file, log files, health URL syntax, Docker CLI, and `lsof`.

### `up`

Start dependency services:

```bash
godesk up <project>
```

If `up_cmd` is configured, godesk runs it as a shell command from the project root. Otherwise it runs Docker Compose with the resolved compose file.

### `ports`

Show port occupancy:

```bash
godesk ports <project>
```

Ports are collected from `.env` values with port-like keys and from Docker Compose published ports.

### `health`

Check configured health URLs:

```bash
godesk health <project>
```

Configure URLs in `.godesk.yaml`:

```yaml
health_urls:
  - http://localhost:8080/health
```

### `logs`

Tail configured log files and Docker Compose service logs:

```bash
godesk logs <project>
```

Tail only file logs:

```bash
godesk logs <project> --files-only --tail 100
```

Tail only Compose logs for specific services:

```bash
godesk logs <project> api worker --compose-only
```

Configure file logs in `.godesk.yaml`:

```yaml
log_files:
  - ./logs/app.log
  - ./logs/worker.log
```

### `lint`

Run the configured lint command:

```bash
godesk lint <project>
```

Configure it in `.godesk.yaml`. The command runs through the system shell from the project root:

```yaml
lint_cmd: "go vet ./... && golangci-lint run"
```

## Development

Project development rules live in:

[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
