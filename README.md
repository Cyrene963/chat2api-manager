# chat2api manager

A single-file Web UI for configuring and testing a local `chat2api` reverse proxy.

## What It Does

- Stores backend profiles in the browser.
- Tests `/v1/accTokens` with a local API key.
- Edits local API keys, direct token prefixes, global proxy settings, and ChatGPT Web account pool entries.
- Imports and exports `app.dev.yaml`.
- Tests `/v1/chat/completions` and `/v1/responses` with a user-specified model such as `gpt-5.5-pro`.
- If the backend includes `/admin/api/config`, it can load and save configuration remotely.

## Usage

Open `index.html` in a browser.

For the local test setup used during development:

```text
Backend URL: http://127.0.0.1:3040
API key: sk-your-local-key
```

For NewAPI/sub2api style integrations, use:

```text
Base URL: http://127.0.0.1:3040/v1
API key: one of auth.access_tokens
Model: gpt-5.5-pro, or another upstream model slug your ChatGPT Web account can access
```

If NewAPI/sub2api runs inside Docker, replace `127.0.0.1` with `host.docker.internal` or the host LAN IP.

## Notes

The upstream ChatGPT Web access token belongs in the account pool, not in NewAPI/sub2api.

The model name is passed through by `chat2api`. This UI does not magically grant access to Pro models; the ChatGPT Web account token must already have the required entitlement.
