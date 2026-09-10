# Vanguard

NTS POS client. Protocol, print, and operator flows are Go; the UI is HTML/CSS/JS in a [Wails](https://wails.io) desktop shell.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/) (Go 1.23.3+ on macOS 15)
- Wails CLI v2.9.2

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Run

```bash
go test ./internal/...
wails dev
```

Or without the watcher:

```bash
wails build
./build/bin/vanguard
```

Place `app.properties` in the working directory (`server`, `port`, `printerName`, …). Copy `app.properties.example` to start. Preferences are stored in the OS config dir as `vanguard/prefs.json` (including `language`: `sr`, `hr`, or `en`).

To talk to Fulgor over HTTP instead of the legacy xrecharge socket, open **Settings → Application** (or set the same keys in `app.properties`):

```
transport=rest
server=127.0.0.1
port=8080
```

The Application screen has host, port, connection (`socket` / `rest`), and HTTP scheme. Default `transport=socket` keeps the old TCP protocol on port 10887.

Window size is 480×560. Keyboard: Enter confirms, Escape goes back, digits 1–9 select menu tiles when shortcuts are enabled.

## Package

macOS / current OS:

```bash
wails build
```

Windows `.exe` (run on Windows, or from this Mac with Docker):

```bash
wails build -platform windows/amd64
```

The binary lands in `build/bin/`.
