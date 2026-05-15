# godesk

中文 | [English](README.md)

godesk 是一个用于本地 Go 后端项目的 CLI 工作台。它可以扫描本地 Go 模块、解析项目配置、读取 `.env` 和 Docker Compose 文件、启动依赖服务、查看端口占用、跟踪日志、检查健康状态，并运行已配置的 lint 命令。

## 功能

- 通过 `go.mod` 扫描本地 Go 项目
- 识别项目附近的 `.env` 和 Docker Compose 文件
- 生成项目级 `.godesk.yaml`
- 管理默认扫描根目录
- 列出已索引项目
- 打开交互式终端工作台
- 查看解析后的项目配置、环境变量和 compose 服务
- 使用 Docker Compose 或自定义命令启动依赖服务
- 从 env 和 compose 配置中查看本地端口占用
- 跟踪已配置的文件日志和 Docker Compose 服务日志
- 检查已配置的健康检查地址
- 运行已配置的 lint 命令

## 安装

直接从源码运行：

```bash
go run ./cmd/godesk --help
```

构建本地二进制：

```bash
go build -o godesk ./cmd/godesk
```

然后运行：

```bash
./godesk --help
```

## 基本流程

扫描包含 Go 项目的目录：

```bash
godesk scan /Users/penty/Desktop/Projects/fuu
```

保存默认扫描根目录：

```bash
godesk roots add /Users/penty/Desktop/Projects/fuu
godesk scan
```

列出已索引项目：

```bash
godesk list
```

打开交互式工作台：

```bash
godesk tui
```

生成项目配置：

```bash
godesk init fzuhelper-server
```

查看解析后的项目：

```bash
godesk inspect fzuhelper-server
```

查看端口：

```bash
godesk ports fzuhelper-server
```

检查健康状态：

```bash
godesk health fzuhelper-server
```

跟踪日志：

```bash
godesk logs fzuhelper-server
```

启动依赖服务：

```bash
godesk up fzuhelper-server
```

运行 lint：

```bash
godesk lint fzuhelper-server
```

## 命令模型

项目级命令使用统一格式：

```bash
godesk <command> <project>
```

当前项目级命令：

```bash
godesk init <project>
godesk inspect <project>
godesk up <project>
godesk ports <project>
godesk health <project>
godesk logs <project> [service...]
godesk lint <project>
```

全局命令：

```bash
godesk roots add <path>
godesk roots list
godesk roots remove <path>
godesk scan [root...]
godesk list
godesk tui
```

对于还没有扫描进索引的项目，可以直接初始化当前 Go 模块：

```bash
godesk init-local
```

## 项目配置

每个项目可以在 Go 模块根目录定义 `.godesk.yaml`：

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

`godesk init <project>` 会为已索引项目创建这个文件。`godesk init-local` 会为当前 Go 模块创建这个文件。

重新扫描时，`.godesk.yaml` 会覆盖自动发现到的项目值。

## 发现规则

godesk 会遍历扫描根目录，寻找包含 `go.mod` 的目录作为 Go 项目。

每个项目会解析这些文件：

```text
.env
docker-compose.yml
docker-compose.yaml
compose.yaml
```

发现逻辑会先从 Go 模块目录向上查到扫描根目录，再从项目根目录向下进行有深度限制的搜索。它支持这样的目录结构：

```text
project/
  go.mod
  docker/
    docker-compose.yml
```

## 命令

### `scan`

扫描根目录并保存项目索引：

```bash
godesk scan /path/to/workspace
```

也支持旧的 flag 形式：

```bash
godesk scan --root /path/to/workspace
```

### `list`

显示已索引项目：

```bash
godesk list
```

### `roots`

管理默认扫描根目录：

```bash
godesk roots add /path/to/workspace
godesk roots list
godesk roots remove /path/to/workspace
```

添加 roots 后，可以直接扫描：

```bash
godesk scan
```

### `tui`

打开交互式项目工作台：

```bash
godesk tui
```

快捷键：

```text
j/k 或方向键   切换项目
r              重新加载项目索引
i              对当前项目运行 inspect
p              对当前项目运行 ports
h              对当前项目运行 health
u              对当前项目运行 up
l              对当前项目运行 logs
q              退出
```

### `init`

为已索引项目创建 `.godesk.yaml`：

```bash
godesk init <project>
```

覆盖已有配置：

```bash
godesk init --force <project>
```

### `init-local`

为当前 Go 模块创建 `.godesk.yaml`：

```bash
godesk init-local
```

### `inspect`

打印项目详情：

```bash
godesk inspect <project>
```

### `up`

启动依赖服务：

```bash
godesk up <project>
```

如果配置了 `up_cmd`，godesk 会在项目根目录把它作为 shell 命令运行。否则会使用解析到的 compose 文件运行 Docker Compose。

### `ports`

显示端口占用：

```bash
godesk ports <project>
```

端口来源包括 `.env` 中类似端口的变量，以及 Docker Compose 中发布的端口。

### `health`

检查已配置的健康检查地址：

```bash
godesk health <project>
```

在 `.godesk.yaml` 中配置：

```yaml
health_urls:
  - http://localhost:8080/health
```

### `logs`

跟踪已配置的文件日志和 Docker Compose 服务日志：

```bash
godesk logs <project>
```

只跟踪文件日志：

```bash
godesk logs <project> --files-only --tail 100
```

只跟踪指定 Compose 服务日志：

```bash
godesk logs <project> api worker --compose-only
```

在 `.godesk.yaml` 中配置文件日志：

```yaml
log_files:
  - ./logs/app.log
  - ./logs/worker.log
```

### `lint`

运行已配置的 lint 命令：

```bash
godesk lint <project>
```

在 `.godesk.yaml` 中配置。命令会在项目根目录通过系统 shell 执行：

```yaml
lint_cmd: "go vet ./... && golangci-lint run"
```

## 开发

项目开发规范位于：

[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
