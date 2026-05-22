# godesk 开发指南

中文 | [English](DEVELOPMENT.md)

## 目标

godesk 是一个本地 Go 后端工作台管理器。它帮助开发者扫描本地 Go 项目、解析项目配置、启动依赖服务、查看端口状态、跟踪日志，并运行已配置的项目工具。

当前产品入口包括 CLI，以及通过 `godesk tui` 打开的无额外依赖终端工作台。

## 环境要求

必需工具：

```text
Go 1.25+
```

运行时集成：

```text
Docker CLI 和 Docker Compose 用于 `godesk up` 和 `godesk logs`
lsof 用于 `godesk ports`
```

常用本地检查：

```bash
go run ./cmd/godesk --help
go run ./cmd/godesk list
docker version
docker compose version
lsof -v
```

## 仓库结构

```text
cmd/godesk/main.go       CLI 入口
internal/cli             Cobra 命令注册、命令处理和 TUI
internal/config          全局配置、项目索引和 .godesk.yaml 处理
internal/project         Go 项目扫描和文件发现
internal/envfile         .env 解析器
internal/compose         Docker Compose 解析器
internal/docker          Docker Compose 进程集成
internal/ports           端口候选提取和 lsof 检查
internal/runner          自定义 shell 命令执行器
docs                     开发文档
```

共享行为放在 internal 包中，CLI 命令和 TUI 动作都使用同一套发现、解析、Docker、端口和命令执行逻辑。

## 命令模型

命令分为项目级命令和全局命令。

项目级命令使用统一格式：

```bash
godesk <command> <project>
```

当前项目级命令：

```bash
godesk init <project>
godesk inspect <project>
godesk doctor <project>
godesk up <project>
godesk ps <project>
godesk ports <project>
godesk health <project>
godesk logs <project> [service...]
godesk lint <project>
```

全局命令操作工作台索引或扫描根目录：

```bash
godesk list
godesk roots add <path>
godesk roots list
godesk roots remove <path>
godesk scan [root...]
godesk tui
```

已索引项目命令使用和现有项目级命令一致的位置接收项目名。

本地直接流程使用独立且明确的命令名。例如：

```bash
godesk init-local
```

## 配置文件

godesk 将全局文件存放在 `os.UserConfigDir()/godesk` 下。

macOS 上通常是：

```text
~/Library/Application Support/godesk
```

当前全局文件：

```text
config.yaml    扫描根目录
projects.yaml  扫描后的项目索引
```

每个 Go 项目都可以在模块根目录定义 `.godesk.yaml`：

```yaml
name: my-project
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

`godesk init <project>` 为已索引项目生成 `.godesk.yaml`。项目名来自索引记录，发现到的路径以项目根目录为基准写入。

`godesk init-local` 会从当前目录向上查找 `go.mod`，并为当前 Go 项目生成 `.godesk.yaml`。

项目重新扫描时，项目配置会覆盖扫描结果。

## 扫描和发现

扫描通过查找包含 `go.mod` 的目录来发现 Go 项目。

项目发现会一起解析这些文件：

```text
.env
docker-compose.yml
docker-compose.yaml
compose.yaml
```

发现逻辑先从 Go 模块目录向上查找到扫描根目录，再从项目根目录向下进行有深度限制的搜索。这支持下面这种布局：

```text
project/
  go.mod
  docker/
    docker-compose.yml
```

当向下搜索存在多个匹配项时，优先选择层级最浅的路径；层级相同时，选择字典序最靠前的路径。

## 运行行为

`up` 会启动项目依赖服务。配置了 `up_cmd` 时会把它作为 shell 命令执行；否则使用解析到的 compose 文件运行 Docker Compose。

`ps` 会基于解析到的 `compose_file` 查看 Docker Compose 服务状态。

`ports` 会从 `.env` 中端口类变量和 Docker Compose 已发布端口收集候选端口，然后报告本地监听状态。

`health` 会检查 `.godesk.yaml` 中的 `health_urls`，并报告状态码、耗时和请求错误。

`logs` 会跟踪已配置的 `log_files` 和 Docker Compose 服务日志。服务参数只过滤 Compose 日志。

`lint` 会在配置了 `lint_cmd` 时把 `.godesk.yaml` 中的命令作为 shell 命令执行。

`inspect` 会打印解析后的项目配置、env 条目和 compose 服务。

`doctor` 会校验项目配置和宿主机工具可用性，不启动服务，也不修改文件。

`tui` 会基于项目索引打开交互式终端工作台。它复用已索引项目数据和现有命令行为，作为统一入口使用。

## 新增项目级命令

作用于单个已索引项目的命令按这个流程添加：

1. 在 `internal/cli` 下新增命令文件，例如 `health.go`。
2. 定义 `newHealthCommand(app *appContext) *cobra.Command`。
3. 将 `Use` 设置为 `health <project>`。
4. 使用 `requireProjectName(cmd, args)` 做参数校验。
5. 使用 `app.store.FindProject(name)` 解析项目。
6. 可复用行为放进职责聚焦的 internal 包。
7. 在 `internal/cli/root.go` 注册命令。
8. 更新 `README.md`、`README.zh-CN.md` 和开发指南。
9. 用 help 输出和一条命令级 smoke 检查验证。

项目命令骨架：

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

## 新增全局命令

全局命令用于操作整体索引、配置或扫描根目录。

1. 在 `internal/cli` 下新增命令文件。
2. 将 `Args` 设置为 `cobra.NoArgs` 或定义明确的位置参数。
3. 通过 `app.store` 读取配置或索引。
4. 列表型输出使用表格形式。
5. 在 `internal/cli/root.go` 注册命令。
6. 更新用户文档和开发文档。

全局命令示例：

```bash
godesk list
godesk roots add <path>
godesk roots list
godesk roots remove <path>
godesk scan [root...]
```

## 新增项目配置字段

当新字段属于 `.godesk.yaml` 时，按这个流程添加：

1. 在 `internal/project/project.go` 的 `project.Project` 中添加字段。
2. 在 `internal/config/config.go` 的 `config.ProjectOverride` 中添加字段。
3. 在 `config.ApplyOverride` 中应用字段。
4. 在 `internal/cli/init.go` 生成 `.godesk.yaml` 时写入字段。
5. 在对应命令或 internal 包中使用字段。
6. 将字段加入 README 示例和两份开发指南。
7. 验证 `godesk init <project>` 或 `godesk init-local` 的输出。

`.godesk.yaml` 中的项目文件路径使用相对项目根目录的路径。

## 新增发现规则

发现逻辑放在 `internal/project`。

新增项目文件类型时，将相关文件放在同一次遍历中发现，保持扫描行为可预测且高效。

向下发现的优先级：

```text
1. 层级最浅的路径
2. 字典序最靠前的路径
```

可能位于子目录的项目本地文件使用有界向下搜索。

## 输出风格

终端输出保持紧凑。

多行资源输出使用 `text/tabwriter`，并提供表头。

项目详情使用稳定标签：

```text
name:
path:
env:
compose:
lint:
up:
health:
logs:
```

缺失的可选值使用 `-`。

## 常见本地流程

扫描并查看工作区：

```bash
go run ./cmd/godesk scan /path/to/workspace
go run ./cmd/godesk list
go run ./cmd/godesk inspect <project>
go run ./cmd/godesk doctor <project>
```

初始化项目配置：

```bash
go run ./cmd/godesk init <project>
go run ./cmd/godesk init-local
```

检查运行状态：

```bash
go run ./cmd/godesk ports <project>
go run ./cmd/godesk up <project>
go run ./cmd/godesk health <project>
go run ./cmd/godesk logs <project> --tail 50
```

运行已配置工具：

```bash
go run ./cmd/godesk lint <project>
```

## 调试

扫描问题对比这些命令输出：

```bash
go run ./cmd/godesk scan /path/to/workspace
go run ./cmd/godesk list
go run ./cmd/godesk inspect <project>
```

发现问题时查看项目布局和 `.godesk.yaml`：

```bash
find /path/to/project -maxdepth 4 -name go.mod -o -name .env -o -name 'docker-compose.yml' -o -name 'docker-compose.yaml' -o -name 'compose.yaml'
```

Docker 问题先直接验证 Docker：

```bash
docker version
docker compose version
docker compose -f /path/to/docker-compose.yml config
```

端口问题直接验证端口：

```bash
lsof -nP -iTCP:<port> -sTCP:LISTEN
```

## 验证

窄范围 CLI 变更使用命令级 smoke 检查：

```bash
go run ./cmd/godesk --help
go run ./cmd/godesk <command> --help
go run ./cmd/godesk list
```

对于会写入用户配置或项目文件的命令，验证生成路径以及输出中的项目、env、compose 解析值。

对于 Docker 行为，先验证 Docker CLI 可用性和 Docker daemon 访问，再定位 godesk 自身行为。

## 当前范围和后续扩展

当前已实现范围：

```text
scan
list
roots
init
init-local
inspect
doctor
up
ps
ports
health
logs
lint
tui
```

自然的后续扩展：

```text
Docker 服务状态视图
全局配置编辑命令
```
