# chat2api manager

中文 | [English](README.en.md)

`chat2api manager` 是一个集成版项目：保留 `aurorax-neo/chat2api` 的 OpenAI 兼容反代能力，并内置一个管理 WebUI，用来配置本地 API key、ChatGPT Web 账号池、代理和模型测试。

它的目标链路是：

```text
NewAPI / sub2api / 其他 OpenAI 兼容客户端
  -> chat2api 本地 API key
  -> chat2api 账号池
  -> ChatGPT Web access token
  -> ChatGPT Web 上游模型
```

## 功能

- `GET /admin`：内置管理页面。
- `GET /admin/api/config`：读取当前 YAML 配置。
- `PUT /admin/api/config`：保存配置，写入 `.bak` 备份并刷新运行中的账号池。
- `index.html`：可单独打开的静态管理页，也可以通过 `/admin` 使用内置版。
- 支持管理 `auth.access_tokens`，这些 key 给 NewAPI/sub2api 使用。
- 支持管理 `chatgpts` 账号池，里面填写 ChatGPT Web 的真实 `access_token`。
- 支持测试 `/v1/chat/completions` 和 `/v1/responses`。
- 支持导入/导出 `app.dev.yaml`。

## 快速启动

### Linux 服务器 Docker 部署

推荐在 Linux 服务器上用 Docker Compose 运行一体版。

```bash
git clone https://github.com/Cyrene963/chat2api-manager.git
cd chat2api-manager
chmod +x scripts/deploy-linux.sh
./scripts/deploy-linux.sh
```

脚本会自动创建：

```text
.chat2api/conf/app.prod.yaml
.chat2api/logs/
```

如果系统安装了 `openssl`，脚本会自动生成一个本地 API key。第一次部署后请编辑：

```bash
nano .chat2api/conf/app.prod.yaml
```

需要重点改：

```yaml
auth:
  access_tokens:
    - sk-your-local-key

chatgpts:
  - access_token: your-chatgpt-web-access-token
    type: pro
```

改完重启：

```bash
docker compose restart
```

默认端口：

```text
3040
```

如果想换宿主机端口：

```bash
CHAT2API_PORT=7846 docker compose up -d --build
```

访问：

```text
WebUI: http://服务器IP:3040/admin
OpenAI Base URL: http://服务器IP:3040/v1
```

服务器安全组/防火墙需要放行对应端口。生产环境建议只对内网开放，或者放在 Nginx/Caddy/Cloudflare Access 等额外鉴权后面。

### 本地源码运行

复制配置模板：

```bash
cp conf/app.demo.yaml conf/app.dev.yaml
```

Windows PowerShell：

```powershell
Copy-Item .\conf\app.demo.yaml .\conf\app.dev.yaml
```

启动服务：

```bash
go run ./cmd
```

默认监听：

```text
http://127.0.0.1:3040
```

打开 WebUI：

```text
http://127.0.0.1:3040/admin
```

如果只想打开静态页面，也可以直接打开仓库根目录的 `index.html`。

## WebUI 怎么用

### 顶部连接区

`Profile`：保存不同后端配置，比如本地、服务器、Docker。

`Backend URL`：chat2api 服务地址。例如：

```text
http://127.0.0.1:3040
```

`API key`：本地管理 key，来自配置里的 `auth.access_tokens`。demo 默认是：

```text
sk-your-local-key
```

`Connect`：检测 `/v1/accTokens` 是否能连通。

`Load Remote`：从后端读取配置。需要后端包含本项目新增的 `/admin/api/config`。

`Save Remote`：保存配置回后端 YAML 文件。保存时会写出 `conf/app.<ENV>.yaml.bak` 备份。

### Server 页面

`Bind`：服务监听地址。本地建议 `127.0.0.1`，服务器部署一般是 `0.0.0.0`。

`Port`：服务端口，默认 `3040`。

`Global proxy`：全局代理。账号级代理为空时会使用这里的代理。

`ChatGPT base URL`：ChatGPT Web 上游地址，默认：

```text
https://chatgpt.com
```

`Local API Keys`：给 NewAPI/sub2api 使用的 key。客户端请求时使用：

```text
Authorization: Bearer <local-api-key>
```

`Direct Token Prefixes`：高级模式。配置私有前缀后，请求方可以直传真实 ChatGPT Web access token，跳过账号池。普通使用不建议开启。

### Accounts 页面

这里配置 ChatGPT Web 账号池。

每个账号最关键的是：

```text
access_token
```

注意它不是 OpenAI API key，也不是 NewAPI/sub2api 里填的 key，而是 ChatGPT Web 的 access token。不要带 `Bearer ` 前缀。

字段说明：

- `Email`：备注用。
- `Type`：备注用，例如 `codex`、`pro`、`plus`。
- `Access token`：ChatGPT Web 上游 token，账号池实际使用它请求上游。
- `Refresh token` / `ID token`：可选辅助字段。
- `Proxy`：账号专属代理，优先级高于全局代理。
- `Expired` / `Last refresh`：备注或人工管理用。

### Test 页面

可以直接测试反代。

`Endpoint`：

```text
/v1/chat/completions
/v1/responses
```

`Model`：会原样传给 ChatGPT Web 上游。这个项目不会做模型名映射。

常见测试值：

```text
auto
gpt-5.5-pro
```

如果请求里的 model 为空，后端会自动使用：

```text
auto
```

`Stream`：是否流式返回。

`Message`：测试消息。

### Raw YAML 页面

直接查看或编辑配置 YAML，适合复制到服务器部署。

## 接入 NewAPI / sub2api

在 NewAPI/sub2api 里添加 OpenAI 兼容渠道：

```text
Base URL: http://服务器IP:3040/v1
API Key: sk-your-local-key
Model: gpt-5.5-pro
```

如果 NewAPI/sub2api 和 chat2api-manager 在同一台 Linux 服务器上，并且 NewAPI/sub2api 不是 Docker 容器，可以用：

```text
http://127.0.0.1:3040/v1
```

如果 NewAPI/sub2api 在 Docker 里运行，`127.0.0.1` 指的是容器自己，需要改成：

```text
http://host.docker.internal:3040/v1
```

Linux Docker 环境里 `host.docker.internal` 不一定默认可用，更稳的是使用宿主机局域网 IP，或者把两个服务放进同一个 Docker network 后使用服务名。

宿主机局域网 IP 示例：

```text
http://192.168.x.x:3040/v1
```

## 关于 Pro 模型和 Extended 思考模式

`model` 只负责选择模型。比如你填：

```text
gpt-5.5-pro
```

后端会把它原样传给 ChatGPT Web 上游。

`Extended` 不是模型名，而是 ChatGPT 里的思考时间/努力程度选项。当前代码没有单独暴露 thinking time 参数。如果上游要求在请求体里传额外字段，后续需要抓取 ChatGPT Web 的真实请求，再把对应字段补进 `chat2api`。

也就是说，本项目可以帮你选择 Pro 模型，但不能凭空给普通账号开通 Pro，也不能保证一定能强制 Extended/Heavy 思考模式。最终取决于账号权限和上游当前接受的请求格式。

## 配置文件

服务读取：

```text
conf/app.<ENV>.yaml
```

默认：

```text
conf/app.dev.yaml
```

配置模板：

```text
conf/app.demo.yaml
```

## 安全提醒

- 不要把 GitHub token、ChatGPT Web access token、OpenAI API key 提交到仓库。
- WebUI 不会回显已保存的真实 token，只显示遮罩值。
- 保存远程配置时会生成 `.bak` 备份。
- 生产环境请放在内网、反代鉴权或防火墙后面。

## 原项目

本项目基于 `aurorax-neo/chat2api` 扩展管理 WebUI 和配置 API。
