# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go HTTP API server implementing the TRMNL device API specification. It serves as a local API server for TRMNL e-ink display devices, handling automatic device authentication, image serving, and logging. The server uses command line arguments for configuration and serves images directly from a specified directory.

## Essential Commands

### Building
- `make build` - Build for current platform
- `make build-all` - Cross-compile for all platforms
- `go build -o trmnl-api-server .` - Direct go build

### Testing & Quality
- `make test` or `go test -v ./...` - Run tests
- `make lint` or `golangci-lint run` - Run linter with comprehensive checks
- `go vet ./...` - Run go vet

### Development
- `make run` or `go run .` - Run the server directly
- `make dev` - Build with race detector enabled
- `make deps` - Download and tidy dependencies

### Configuration
- Command line arguments: `--port`, `--bind`, `--mac`, `--help`
- Positional argument: image directory path (defaults to current directory)
- **Required** environment variable: `SECRET_KEY_BASE` for API key generation
- Default server runs on port 3000, binding to all interfaces (0.0.0.0)
- Optional MAC address whitelist via `--mac` (comma-separated, case insensitive)
- Use `./trmnl-api-server --help` to see all options

## Architecture

### Core Components

**Single-file application** (`trmnl_api_server.go`) with these key structures:

- **Server struct**: Main server with config and device state management
- **Device registration**: Auto-authentication of all MAC addresses (or whitelist via --mac) with SHA1 API key generation
- **Image serving**: Incremental image delivery from shared image directory
- **State management**: Thread-safe per-device playlist position tracking with mutex
- **Direct image endpoint**: `/images/{filename}` with security validation
- **Middleware**: CORS and logging middleware for all requests

### API Endpoints

- `GET /api/setup` - Device registration (requires `ID` header with MAC address) - **Auto-authenticates all devices or uses whitelist**
- `GET /api/display` - Image serving (requires `Access-Token` header) - Returns next image in playlist
- `POST /api/log` - Device logging (requires `Access-Token` header)
- `GET /images/{filename}` - Direct image serving with security validation

### Image Directory Structure

Single shared image directory containing:
- `.bmp` or `.png` image files
- Optional duration control via `--XX` suffix (e.g., `image--30.bmp` for 30 seconds)
- Images served in lexicographic (alphabetical) order
- Each device maintains its own position in the playlist

### Key Features

- **Incremental image serving**: Cycles through images in lexicographic order
- **Per-device state tracking**: Each device maintains independent playlist position
- **Duration parsing**: Extracts custom durations from filenames
- **Dynamic URL generation**: Uses request hostname for image URLs (works with any domain/IP)
- **Friendly ID format**: ABC-123 (3 chars, dash, 3 chars)
- **Authentication logging**: Logs MAC address authentication attempts to terminal
- **Auto-authentication**: All MAC addresses are automatically authenticated (or whitelist via --mac option)
- **Thread-safe state**: Concurrent device state management
- **Security validation**: Path traversal protection for image serving
- **Cross-platform builds**: Linux, macOS, Windows (AMD64 & ARM64)

## Dependencies

- `github.com/gorilla/mux` - HTTP routing
- Go 1.21+ required (no external config dependencies)

## Build System

Uses Makefile with comprehensive targets including cross-compilation, distribution packaging, and checksums. Build includes version info via ldflags.