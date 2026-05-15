# godesk

[中文](README.zh-CN.md) | English

godesk is a CLI workspace manager for local Go backend projects. It scans local Go modules, resolves project config, reads `.env` and Docker Compose files, starts dependency services, checks port occupancy, checks health URLs, and runs configured lint commands.

## Features

- Scan local directories for Go projects by `go.mod`
- Resolve `.env` and Docker Compose files near a project
- Generate per-project `.godesk.yaml`
- List indexed projects
- Inspect resolved project config, env entries, and compose services
- Start dependency services with Docker Compose or a custom command
- Show local port occupancy from env and compose config
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

List indexed projects:

```bash
godesk list
```

Generate project config:

```bash
godesk init fzuhelper-server
```

Inspect the resolved project:

```bash
godesk inspect fzuhelper-server
```

Check ports:

```bash
godesk ports fzuhelper-server
```

Check health:

```bash
godesk health fzuhelper-server
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
godesk up <project>
godesk ports <project>
godesk health <project>
godesk lint <project>
```

Global commands:

```bash
godesk scan [root...]
godesk list
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
lint_cmd: golangci-lint run
up_cmd: make up
health_urls:
  - http://localhost:8080/health
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

### `up`

Start dependency services:

```bash
godesk up <project>
```

If `up_cmd` is configured, godesk runs it from the project root. Otherwise it runs Docker Compose with the resolved compose file.

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

### `lint`

Run the configured lint command:

```bash
godesk lint <project>
```

Configure it in `.godesk.yaml`:

```yaml
lint_cmd: golangci-lint run
```

## Development

Project development rules live in:

[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
