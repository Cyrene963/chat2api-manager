# chat2api manager

[中文](README.md) | English

`chat2api manager` is an integrated build of `aurorax-neo/chat2api` with a management WebUI for local API keys, ChatGPT Web account-pool tokens, proxy settings, and model testing.

Target flow:

```text
NewAPI / sub2api / OpenAI-compatible client
  -> local chat2api API key
  -> chat2api account pool
  -> ChatGPT Web access token
  -> ChatGPT Web upstream model
```

## Features

- `GET /admin`: built-in management UI.
- `GET /admin/api/config`: read the active YAML config.
- `PUT /admin/api/config`: save config, create a `.bak` backup, and refresh the running account pool.
- `index.html`: standalone static manager page.
- Manage `auth.access_tokens` for NewAPI/sub2api.
- Manage `chatgpts` account-pool entries with real ChatGPT Web `access_token` values.
- Test `/v1/chat/completions` and `/v1/responses`.
- Optionally proxy OpenAI Responses / deep research when `OPENAI_API_KEY` is set.
- Import and export `app.dev.yaml`.

## Quick Start

### Linux Server Docker Deployment

Docker Compose is the recommended deployment path for Linux servers.

```bash
git clone https://github.com/Cyrene963/chat2api-manager.git
cd chat2api-manager
chmod +x scripts/deploy-linux.sh
./scripts/deploy-linux.sh
```

The script creates:

```text
.chat2api/conf/app.prod.yaml
.chat2api/logs/
```

If `openssl` is available, it also generates a local API key. After the first run, edit:

```bash
nano .chat2api/conf/app.prod.yaml
```

Important fields:

```yaml
auth:
  access_tokens:
    - sk-your-local-key

chatgpts:
  - access_token: your-chatgpt-web-access-token
    type: pro
```

Restart after editing:

```bash
docker compose restart
```

Default port:

```text
3040
```

Use a different host port:

```bash
CHAT2API_PORT=7846 docker compose up -d --build
```

URLs:

```text
WebUI: http://SERVER_IP:3040/admin
OpenAI Base URL: http://SERVER_IP:3040/v1
```

Open the port in your firewall/security group. In production, expose it only to a private network or place it behind Nginx/Caddy/Cloudflare Access or another authentication layer.

### Local Source Run

Copy the demo config:

```bash
cp conf/app.demo.yaml conf/app.dev.yaml
```

Windows PowerShell:

```powershell
Copy-Item .\conf\app.demo.yaml .\conf\app.dev.yaml
```

Run:

```bash
go run ./cmd
```

Default server:

```text
http://127.0.0.1:3040
```

Open the WebUI:

```text
http://127.0.0.1:3040/admin
```

You can also open the root `index.html` directly as a standalone static page.

## WebUI Guide

### Top Connection Area

`Profile`: save multiple backend profiles.

`Backend URL`: the chat2api backend URL, for example:

```text
http://127.0.0.1:3040
```

`API key`: a local management key from `auth.access_tokens`.

`Connect`: checks `/v1/accTokens`.

`Load Remote`: reads config from the backend.

`Save Remote`: writes config back to the backend YAML file and creates a `.bak` backup.

### Server Page

`Bind`: server bind address.

`Port`: server port.

`Global proxy`: fallback proxy for accounts without account-level proxy.

`ChatGPT base URL`: defaults to `https://chatgpt.com`.

`Local API Keys`: keys used by NewAPI/sub2api.

`Direct Token Prefixes`: advanced direct-token mode. Usually leave it disabled.

### Accounts Page

This is the ChatGPT Web account pool.

The required field is:

```text
access_token
```

It is a ChatGPT Web access token, not an OpenAI API key and not the key used by NewAPI/sub2api. Do not include the `Bearer ` prefix.

Account-level proxy takes precedence over the global proxy.

### Test Page

Test OpenAI-compatible requests.

Endpoints:

```text
/v1/chat/completions
/v1/responses
```

`Model` is passed through to ChatGPT Web. This project does not map model names.

Example values:

```text
auto
gpt-5.5-pro
```

If the model is empty, the backend uses `auto`.

### Raw YAML Page

View or edit the generated YAML config directly.

## NewAPI / sub2api Integration

Use an OpenAI-compatible channel:

```text
Base URL: http://SERVER_IP:3040/v1
API Key: sk-your-local-key
Model: gpt-5.5-pro
```

If NewAPI/sub2api runs on the same Linux host and is not containerized, you can use:

```text
http://127.0.0.1:3040/v1
```

If NewAPI/sub2api runs in Docker, `127.0.0.1` points to the container itself. You can try:

```text
http://host.docker.internal:3040/v1
```

On Linux Docker, `host.docker.internal` may not be enabled by default. The host LAN IP or a shared Docker network service name is usually more reliable.

## Pro Models And Extended Thinking

`model` selects the model and is passed through unchanged.

`Extended` is not a model name. It is a ChatGPT thinking-time / effort option. This code does not currently expose a separate thinking-time parameter. If ChatGPT Web requires an extra request field for Extended or Heavy thinking, the next step is to capture the real ChatGPT Web request and add that field to `chat2api`.

The account token must already have the required Pro entitlement.

## OpenAI Deep Research

If you set `OPENAI_API_KEY` or `openai_api_key`, `/v1/responses` can proxy directly to the OpenAI Responses API.

Use the official deep research models:

```text
o3-deep-research
o4-mini-deep-research
```

Typical request shape:

```json
{
  "model": "o3-deep-research",
  "background": true,
  "reasoning": { "summary": "auto" },
  "tools": [{ "type": "web_search_preview" }]
}
```

The backend also exposes `GET /v1/responses/{id}` so you can poll background jobs.

## Security

- Do not commit GitHub tokens, ChatGPT Web access tokens, or OpenAI API keys.
- The WebUI masks saved tokens.
- Remote saves create a `.bak` backup.
- Put the admin UI behind a firewall, private network, or extra authentication in production.

## Upstream

Based on `aurorax-neo/chat2api`, extended with a management WebUI and config API.
