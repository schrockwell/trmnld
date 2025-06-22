# TRMNL API Server

A single-file standalone HTTP server that serves up a directory of images to TRMNL e-ink displays.

## Features

- **Device Setup**: Automatic MAC address authentication with API key generation
- **Image Display**: Incremental image serving with custom duration support
- **Device State Tracking**: Per-device playlist position management in memory
- **Direct Image Serving**: Built-in image endpoint with security validation
- **Logging**: Device log collection with stdout output
- **CORS Support**: Cross-origin request handling for web interfaces
- **Cross-Platform**: Builds for Linux, macOS, and Windows (AMD64 & ARM64)

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
```

### Configuration

**Required Environment Variable:**
```bash
export SECRET_KEY_BASE="your-secret-key-here"
```

**Command Line Options:**
Use `--help` to see all available options:
```bash
./trmnld --help
```

Basic usage:
```bash
# Run with default settings (port 3000, bind to all interfaces)
./trmnld

# Specify custom port and bind address
./trmnld --port 8080 --bind 127.0.0.1

# Use custom image directory
./trmnld /path/to/images

# Enable MAC address whitelist
./trmnld --mac AA:BB:CC:DD:EE:FF,11:22:33:44:55:66

# Combine options
./trmnld --port 8080 --bind 0.0.0.0 --mac AA:BB:CC:DD:EE:FF /path/to/images
```

## API Endpoints

### `GET /api/setup`
Device registration endpoint. **All MAC addresses are authenticated by default, unless --mac whitelist is specified.**

**Headers:**
- `ID`: Device MAC address (e.g., `AA:BB:CC:DD:EE:FF`)

**Response:**
```json
{
  "status": 200,
  "api_key": "generated-api-key",
  "friendly_id": "ABC-123",
  "message": "Register at usetrmnl.com/signup with Device ID 'ABC-123'"
}
```

### `GET /api/display`
Image serving endpoint. Returns the next image in the playlist for each device.

**Headers:**
- `Access-Token`: Device API key
- `Battery-Voltage`: Battery voltage (optional)
- `FW-Version`: Firmware version (optional)
- `RSSI`: Signal strength (optional)
- `Height`: Screen height (optional)
- `Width`: Screen width (optional)

**Response:**
```json
{
  "status": 200,
  "image_url": "http://your-server.com/images/image.bmp",
  "filename": "image.bmp",
  "refresh_rate": 900
}
```

**Note:** The `image_url` uses the hostname from the HTTP request, so it works automatically with any domain/IP.

### `POST /api/log`
Logging endpoint.

**Headers:**
- `Access-Token`: Device API key

**Body:**
```json
{
  "log": "any JSON value - string, object, array, etc."
}
```

### `GET /images/{filename}`
Direct image serving endpoint with security validation.

**Example:**
```
GET /images/myimage.bmp
```

Returns the image file with appropriate content headers.

## Image Duration Control

Images are served with a default duration of 900 seconds. To specify custom durations, append `--XX` to the filename (before the extension):

- `image.bmp` → 900 seconds (default)
- `image--30.bmp` → 30 seconds
- `image--600.png` → 600 seconds (10 minutes)

## Device State Management

Each device maintains its own position in the image playlist:
- Devices cycle through images in lexicographic (alphabetical) order
- Each device's current position is tracked independently in memory
- When a device reaches the last image, it loops back to the first

## MAC Address Authentication

By default, all MAC addresses are automatically authenticated. You can enable a whitelist using the `--mac` option:

```bash
# Allow only specific MAC addresses (case insensitive)
./trmnld --mac AA:BB:CC:DD:EE:FF,11:22:33:44:55:66
```

When a device attempts to register:
- **Allowed**: `MAC address AA:BB:CC:DD:EE:FF was authenticated`
- **Denied**: `MAC address XX:XX:XX:XX:XX:XX was denied`

The friendly ID format is ABC-123 (3 characters, dash, 3 characters).

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

## Command Line Options

| Option              | Type   | Default   | Description                                                                                                        |
| ------------------- | ------ | --------- | ------------------------------------------------------------------------------------------------------------------ |
| `--port`            | int    | `3000`    | HTTP server port                                                                                                   |
| `--bind`            | string | `0.0.0.0` | Address to bind to (0.0.0.0 for all interfaces)                                                                    |
| `--mac`             | string | `""`      | Comma-separated list of allowed MAC addresses (case insensitive). If not specified, all MAC addresses are allowed. |
| `--help`            | bool   | `false`   | Show help message and exit                                                                                         |
| `[image-directory]` | string | `.`       | Directory containing images (positional argument)                                                                  |

## Environment Variables

| Variable          | Required | Description                                                          |
| ----------------- | -------- | -------------------------------------------------------------------- |
| `SECRET_KEY_BASE` | **Yes**  | Secret key for API key generation (server will not start without it) |

## License

[Add your license information here]

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run tests and linting
6. Submit a pull request