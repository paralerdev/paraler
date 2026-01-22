# paraler

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

TUI for running your local dev servers. All of them.

All your projects. All your services. One terminal. Start working — everything's up. Close paraler — everything stops.

## Install

### macOS

```bash
# Homebrew (coming soon)
# brew install paralerdev/tap/paraler

# Or download binary
curl -fsSL https://github.com/paralerdev/paraler/releases/latest/download/paraler-darwin-arm64 -o paraler
chmod +x paraler
sudo mv paraler /usr/local/bin/
```

### Linux

```bash
curl -fsSL https://github.com/paralerdev/paraler/releases/latest/download/paraler-linux-amd64 -o paraler
chmod +x paraler
sudo mv paraler /usr/local/bin/
```

### From source

```bash
go install github.com/paralerdev/paraler/cmd/paraler@latest
```

## Quick Start

```bash
# Scan and add a project
paraler add ~/projects/myapp

# Run
paraler
```

Or create `paraler.yaml` manually:

```yaml
projects:
  myapp:
    path: ~/projects/myapp
    services:
      api:
        cmd: npm run start:dev
        port: 3010
        cwd: ./backend
      web:
        cmd: npm run dev
        port: 5173
        cwd: ./frontend
```

## Features

- **Start/stop/restart** — single keypress, or all at once
- **Logs** — stdout/stderr from every service, with filtering
- **Health checks** — HTTP endpoints and port monitoring
- **Auto-restart** — crashed service comes back automatically
- **Auto-discovery** — detects NestJS, React, Vue, Go, and more
- **Dependencies** — start backend before frontend
- **Hot reload** — edit config, Ctrl+R, no restart needed

## Keybindings

```
Navigation  ↑/k up │ ↓/j down │ Tab switch panel
Services    s start │ x stop │ r restart
Bulk        S start all │ X stop all │ v select
Logs        / filter │ c clear │ e export
Other       a add project │ ? help │ q quit
```

## Config Options

| Field | Description |
|-------|-------------|
| `cmd` | Command to run |
| `cwd` | Working directory (relative to project path) |
| `port` | Port to monitor |
| `health` | HTTP health check URL |
| `env` | Environment variables |
| `depends_on` | Start after these services |
| `auto_restart` | Restart on crash (default: false) |
| `color` | Custom color (hex) |

## Supported Frameworks

Auto-discovery works with:

**Backend:** NestJS, Express, Fastify, Go, Rust, Python
**Frontend:** React, Vue, Svelte, Next.js, Nuxt

> Flutter requires manual config — device selection is interactive.

## Requirements

- Go 1.21+
- macOS or Linux

## License

[MIT](LICENSE)
