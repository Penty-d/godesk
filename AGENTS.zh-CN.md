# Repository Guidelines

中文 | [English](AGENTS.md)

## 项目结构与模块组织

`cmd/godesk/main.go` 是 CLI 入口。Cobra 命令注册、处理和无额外依赖的 TUI 位于 `internal/cli`。共享逻辑按职责拆分：`internal/config` 处理全局配置、扫描根目录、项目索引和 `.godesk.yaml`；`internal/project` 处理 Go 模块扫描和文件发现；`internal/envfile` 与 `internal/compose` 负责解析项目文件；`internal/docker`、`internal/logtail`、`internal/ports`、`internal/runner` 封装运行时集成。用户文档位于 `README.md` 和 `README.zh-CN.md`；开发文档位于 `docs/`。

## 构建、验证和开发命令

在仓库根目录使用这些命令：

```bash
go run ./cmd/godesk --help        # 从源码运行 CLI
go run ./cmd/godesk <cmd> --help  # 查看单个命令
go build -o godesk ./cmd/godesk   # 构建本地二进制
go run ./cmd/godesk list          # smoke 检查配置和索引读取
```

Docker 相关改动使用主机工具直接验证：

```bash
docker version
docker compose version
lsof -v
```

## 代码风格与命名约定

Go 文件使用 `gofmt` 格式化。命令构造函数命名为 `new<Name>Command`，并在 `internal/cli/root.go` 注册。项目级命令使用 `godesk <command> <project>`，并通过 `requireProjectName` 校验参数。全局配置命令使用明确子命令，例如 `godesk roots add <path>`。校验类命令统一输出 `ok`、`warn` 和 `fail`。可复用逻辑放入职责聚焦的 internal 包。

## 验证指南

当前验证以命令级 smoke 检查为主。CLI 改动运行 `go run ./cmd/godesk --help` 和对应的 `go run ./cmd/godesk <command> --help`。会写文件或用户配置的命令需要验证输出路径和解析后的项目值。Docker 与端口行为先确认 Docker 或 `lsof` 本身表现。

## 提交与 PR 指南

现有提交使用简洁的 conventional 风格主题，例如 `init: scaffold godesk project` 和 `refactor: add init commands, remove test cmd, improve project discovery`。提交信息优先使用 `type: summary`。PR 应说明面向用户的命令行为、列出涉及的包、附上 smoke 检查命令，并标明配置文件变化。

## 配置提示

项目配置放在 Go 模块根目录的 `.godesk.yaml`。项目文件路径使用相对模块根目录的路径，例如 `compose_file: docker/docker-compose.yml` 和 `log_files: ["./logs/app.log"]`。可运行命令按 shell 命令保存，例如 `lint_cmd: "go vet ./... && golangci-lint run"`。
