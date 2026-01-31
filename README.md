# Rofi Chrome Tab

rofi-chrome-tab is a Chrome extension to select tabs from rofi window switcher.

## Features

- Switch between Chrome/Chromium tabs using rofi
- Fast keyboard-driven tab selection
- Works with i3, Sway, and other window managers
- Real-time tab list updates when tabs are activated

## Installation

### 1. Install Dependencies

```bash
# Install required tools
sudo apt install rofi netcat-openbsd golang

# For non-i3 window managers, optionally install:
sudo apt install wmctrl  # For general window manager support
```

### 2. Build the Native Messaging Host

```bash
cd go
go build -o rofi-chrome-tab
```

### 3. Configure Chrome Native Messaging

Create the native messaging host configuration file:

```bash
# Create the directory if it doesn't exist
mkdir -p ~/.config/google-chrome/NativeMessagingHosts
# Or for Chromium:
# mkdir -p ~/.config/chromium/NativeMessagingHosts

# Copy the sample config
cp rofi_chrome_tab.sample.json ~/.config/google-chrome/NativeMessagingHosts/rofi_chrome_tab.json
```

Edit `~/.config/google-chrome/NativeMessagingHosts/rofi_chrome_tab.json`:
1. Replace `/PATH/TO/rofi-chrome-tab/go/rofi-chrome-tab` with the absolute path to the compiled binary
2. Replace `CHROME_EXTENSION_ID` with the actual extension ID after installation (see below)

### 4. Install the Chrome Extension

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode" (toggle in top-right corner)
3. Click "Load unpacked"
4. Select the `rofi-chrome-tab` directory (the one containing `manifest.json`)
5. Copy the extension ID shown on the extension card
6. Update the `allowed_origins` in `rofi_chrome_tab.json` with this ID

### 5. Set Up the Rofi Script

```bash
# Make the script executable
chmod +x scripts/rofi-chrome-tab

# Add it to your PATH or create a symlink
sudo ln -s "$(pwd)/scripts/rofi-chrome-tab" /usr/local/bin/rofi-chrome-tab
```

## Usage

Run the rofi tab switcher:

```bash
rofi -show chrome-tab -modi "chrome-tab:rofi-chrome-tab"
```

Or add a keyboard shortcut in your window manager config:

For i3, add to `~/.config/i3/config`:
```
bindsym $mod+t exec rofi -show chrome-tab -modi "chrome-tab:rofi-chrome-tab"
```

For Sway, add to `~/.config/sway/config`:
```
bindsym $mod+t exec rofi -show chrome-tab -modi "chrome-tab:rofi-chrome-tab"
```

## How It Works

1. The Chrome extension communicates with the Go native messaging host via stdin/stdout
2. The Go application creates a Unix socket (e.g., `/tmp/native-app.<pid>.sock`)
3. The rofi script queries tabs via the socket and displays them
4. When you select a tab, rofi sends a selection command back to activate it

## Troubleshooting

### Extension not connecting

1. Check the native messaging host path in `rofi_chrome_tab.json` is correct
2. Ensure the `rofi-chrome-tab` binary is executable: `chmod +x go/rofi-chrome-tab`
3. Verify the extension ID in `allowed_origins` matches the installed extension

### Enable debug logging

Create a debug flag file:
```bash
touch /tmp/.rofi-chrome-tab.debug
```

Then check the logs:
```bash
tail -f /tmp/rofi-chrome-tab.log
```

### No tabs showing in rofi

1. Ensure the Chrome extension is loaded and enabled
2. Check that the native messaging host is running:
   ```bash
   ls -la /tmp/native-app*.sock
   ```
3. Try restarting Chrome

## Development

### Running Tests

```bash
cd go
go test -v ./...
```

### Code Formatting

```bash
cd go
gofmt -s -w .
go vet ./...
```

## License

See LICENSE file for details.
