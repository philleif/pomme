# AGENT.md

## Project Overview
Pomme is a macOS pomodoro timer written in Go with Messages app blocking, menu bar integration, and TUI controls.

## Tech Stack
- **Language**: Go 1.23+
- **TUI**: Bubble Tea (charmbracelet/bubbletea)
- **Styling**: Lipgloss (charmbracelet/lipgloss)
- **Menu Bar**: fyne.io/systray
- **Database**: SQLite (mattn/go-sqlite3)
- **MCP**: mark3labs/mcp-go

## Project Structure
```
cmd/pomme/       - Main entry point
internal/
  blocker/       - Messages.app blocking logic
  client/        - Unix socket client for IPC
  config/        - Configuration management (~/.pomme/config.json)
  daemon/        - Background daemon with socket server
  mcp/           - MCP server implementation
  menubar/       - macOS menu bar integration
  sparkline/     - Braille and Kitty graphics sparklines
  storage/       - SQLite database operations
  timer/         - Pomodoro timer logic
  tui/           - Terminal UI with Bubble Tea
```

## Build Commands
- `make build` - Build binary to ./build/pomme
- `make install` - Build and install to /usr/local/bin
- `make test` - Run tests
- `make clean` - Remove build artifacts and socket file
- `make deps` - Tidy and download dependencies

## Architecture
- **Daemon**: Runs persistently with menu bar icon, exposes Unix socket at ~/.pomme/pomme.sock
- **Client**: CLI commands and TUI communicate with daemon via socket
- **MCP Server**: Runs via `--mcp` flag for AI assistant integration

## Data Locations
- Config: `~/.pomme/config.json`
- Database: `~/.pomme/pomme.db`
- Socket: `~/.pomme/pomme.sock`
