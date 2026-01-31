# Rofi Chrome Tab

rofi-chrome-tab is a chrome extension to select tabs from rofi window switcher.

## Features

- **List all tabs**: View all Chrome tabs from all windows in rofi
- **Switch to tab**: Select a tab to bring it to focus (automatically switches to the correct window)
- **Close tab**: Press Shift+Delete in rofi to close a tab without switching to it
- **Tab information**: Shows tab title, hostname, and full URL for each tab
- **Multi-window support**: Displays tabs from all Chrome windows with window ID information

## Commands

The native messaging host supports the following commands:

- `list` - List all tabs with their IDs, window IDs, hosts, URLs, and titles
- `select <tab_id>` - Switch to and focus the specified tab
- `close <tab_id>` - Close the specified tab

## Tab List Format

Each tab is listed in CSV format:
```
<process_id>,<tab_id>,<window_id>,<hostname>,<url>,<title>
```

## Usage

Run rofi with the script to list and select tabs:
```bash
rofi -modi "chrome:scripts/rofi-chrome-tab" -show chrome
```

- Press Enter to switch to a tab
- Press Shift+Delete to close a tab

