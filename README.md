# Pomme 🍅

A macOS pomodoro timer with Messages app blocking, menu bar integration, and TUI controls. Designed for seamless tmux integration.

## Features

- **Pomodoro Timer**: 25-min work / 5-min break / 20-min long break (after 4 intervals)
- **App Blocking**: Automatically blocks Messages.app during focus intervals
- **Menu Bar**: Live timer display with weekly sparkline
- **TUI Interface**: Compact terminal UI built with Bubble Tea
- **tmux Integration**: Status line output and keybinding commands
- **Tufte-inspired Sparklines**: Braille characters or Kitty graphics (for Ghostty)

## Installation

```bash
# Build
go build -o pomme ./cmd/pomme

# Install to PATH
sudo cp pomme /usr/local/bin/
# or
go install ./cmd/pomme
```

## Usage

### Quick Start

```bash
pomme             # Opens TUI and auto-starts daemon in background
```

The daemon provides the menu bar icon and runs persistently. It auto-starts when you run any pomme command.

### TUI Controls

Controls:
- `s` - Start timer
- `p` - Pause timer
- `k` - Skip to next phase
- `r` - Reset timer
- `b` - Toggle Messages blocking
- `a` - Toggle "always block" mode
- `q` - Quit TUI

### Command Line

```bash
pomme --status        # Print status line (for tmux)
pomme --start         # Start/resume timer
pomme --pause         # Pause timer
pomme --skip          # Skip to next phase
pomme --reset         # Reset timer
pomme --toggle-block  # Toggle Messages blocking
pomme --stats         # Print today's stats with braille sparkline
pomme --graph         # Show pixel-based sparkline (Kitty graphics for Ghostty)
```

## tmux Integration

### Status Bar

Add to your `~/.tmux.conf`:

```bash
# Show pomme status in right status bar
set -g status-right "#(pomme --status) | %H:%M"

# Refresh every 5 seconds
set -g status-interval 5
```

Output example: `🍅 18:32 ▃▅▇▆▄▂█`

### Keybindings

```bash
# Add to ~/.tmux.conf
bind-key P run-shell "pomme --start"
bind-key O run-shell "pomme --pause" 
bind-key K run-shell "pomme --skip"
```

## Data Storage

- Database: `~/.pomme/pomme.db`
- Socket: `~/.pomme/pomme.sock`

## Pomodoro Best Practices

Based on research:

- **25-minute work intervals** maximize focus without burnout
- **5-minute breaks** allow mental recovery
- **Long break (20 min) after 4 intervals** prevents fatigue
- **12 intervals/day** = ~5 hours of deep focused work
- **Block distractions** - Messages blocking prevents context switching

## Sparkline Display

Pomme uses **Tufte-inspired sparklines** to show your weekly progress. Day labels are dynamic based on today's day of week.

### Braille Sparkline (default)
Uses Unicode braille characters for smooth 8-level resolution:
```
Week:  ⣀  ⣤  ⣶  ⣿  ⣷  ⣄  ⣿
       S  M  T  W  T  F  S
       2  5  9  12 10 4  8
                        ↑today
```

### Kitty Graphics (for Ghostty)
Use `--graph` flag for pixel-based sparklines using the Kitty graphics protocol:
```bash
pomme --graph
```
This renders actual pixel graphics in terminals that support it (Ghostty, Kitty).

### Status Line (for tmux)
Compact format with subscript digits:
```
🍅 18:32 ⣀ ⣤ ⣶ ⣿ ⣷ ⣄ ⣿₈
```

## License

MIT
