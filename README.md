# TRMNL API Server

A single-file standalone HTTP server that serves up a directory of images to TRMNL e-ink displays.

## Goals

- **Single Responsibility**: Handles device provisioning and image display, nothing else.
- **Secure**: Provisioning is simple but secure, and images are locked-down from unauthorized clients.
- **Fileless**: No configuration or state management files are written to disk.
- **Cross-Platform**: Binaries are available for Linux, macOS, and Windows.

## Quick Start

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/schrockwell/trmnld/releases).

### Build from Source

**Prerequisites:**
- Go 1.21 or later
- Make (optional, for using Makefile)

**Clone and build:**
```bash
git clone https://github.com/schrockwell/trmnld.git
cd trmnld
make build
# or
go build -o trmnld .
# or
go run trmnld.go
```


### Provision a New Device

1. Follow the directions to [unpair a device](https://help.usetrmnl.com/en/articles/11546838-how-to-unpair-re-pair-a-device).
2. Export the `SECRET_KEY_BASE` environment variable.
3. Start trmnld with the `--setup` option to enable provisioning.
4. Connect to the "TRMNL" Wi-Fi network on your phone to configure the device.
    1. Provide SSID and password.
    2. Tap "Custom Server" to enable BYOS support.
    3. Enter the full hostname to the trmnld server without the trailing slash, e.g. `http://<your-ip>:3000`
    4. Click "Connect".
5. Watch the trmnld logs for `GET /api/setup` followed by `GET /api/display` and `GET /images/...`
6. Restart trmnld, omitting the `--setup` option to prevent unauthorized devices from connecting.

## Command Line Options

| Option              | Type   | Default   | Description                                       |
| ------------------- | ------ | --------- | ------------------------------------------------- |
| `--port`            | int    | `3000`    | HTTP server port                                  |
| `--bind`            | string | `0.0.0.0` | Address to bind to (0.0.0.0 for all interfaces)   |
| `--setup`           | bool   | `false`   | Enable provisioning mode for new devices          |
| `--help`            | bool   | `false`   | Show help message and exit                        |
| `[image-directory]` | string | `.`       | Directory containing images (positional argument) |

## Environment Variables

| Variable          | Required | Description                                     |
| ----------------- | -------- | ----------------------------------------------- |
| `SECRET_KEY_BASE` | **Yes**  | Secret key for API key generation (recommended) |


## Image Duration Control

Images are served with a default duration of 15 minutes. To specify custom durations, append `--XX` to the filename (before the extension):

- `image.bmp` → 900 seconds (default)
- `image--30.bmp` → 30 seconds
- `image--600.png` → 600 seconds (10 minutes)

## Development

### Make Targets

```bash
make help              # Show all available targets
make build             # Build for current platform
make build-all         # Cross-compile for all platforms
make test              # Run tests
make lint              # Run linter
make dev               # Build with race detector
make run               # Run the application
make dist              # Create distribution packages
make clean             # Clean build artifacts
```

### Cross-Compilation

Build for specific platforms:
```bash
make build-linux       # Linux (amd64, arm64)
make build-darwin      # macOS (amd64, arm64)
make build-windows     # Windows (amd64, arm64)
```

### Running Tests

```bash
go test -v ./...
# or
make test
```

### Code Quality

```bash
go vet ./...
golangci-lint run
# or
make lint
```

## GitHub Actions

The repository includes automated CI/CD workflows:

- **CI Pipeline** (`.github/workflows/ci.yml`):
  - Runs tests and linting on every push/PR
  - Cross-compiles for all platforms
  - Uploads build artifacts

- **Release Pipeline** (`.github/workflows/release.yml`):
  - Triggers on git tags (`v*`)
  - Creates GitHub releases with binaries
  - Generates checksums and release notes

### Creating a Release

```bash
git tag v1.0.0
git push origin v1.0.0
```