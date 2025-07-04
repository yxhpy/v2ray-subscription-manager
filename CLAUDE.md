# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a V2Ray subscription manager with Web UI - a high-performance proxy management tool that supports multiple protocols (VLESS, Shadowsocks, Hysteria2), batch speed testing, and intelligent proxy switching. The project uses a dual-architecture approach with both CLI tools and a modern Web UI.

## Build Commands

### Main CLI Application
```bash
# Build all platforms
./scripts/build.sh

# Build current platform only
go build -o bin/v2ray-subscription-manager ./cmd/v2ray-manager/

# Build with version
./scripts/build.sh v1.3.0
```

### Web UI
```bash
# Build Web UI for all platforms
./scripts/build_webui.sh

# Build Web UI for current platform
go build -o web-ui ./cmd/web-ui/

# Quick start Web UI
./start_web_ui.sh
```

### Cleanup Tool
```bash
# Build cleanup tool
./scripts/build_cleanup.sh
go build -o bin/cleanup ./cmd/cleanup.go
```

## Development Commands

### Running the Application
```bash
# Start Web UI (recommended)
./start_web_ui.sh
# Or: go run ./cmd/web-ui/main.go

# CLI usage examples
./bin/v2ray-subscription-manager parse <subscription-url>
./bin/v2ray-subscription-manager speed-test <subscription-url>
./bin/v2ray-subscription-manager start-proxy random <subscription-url>
```

### Testing
```bash
# Run integration tests
./test_all_features.sh

# Test specific components
go test ./internal/...
go test ./pkg/...

# Test port conflict handling (Web UI)
# Open test_port_conflict_advanced.html in browser
# Test database state consistency and conflict resolution
```

### Dependencies
```bash
# Install Go dependencies
go mod tidy

# Download V2Ray core
./bin/v2ray-subscription-manager download-v2ray

# Download Hysteria2 client
./bin/v2ray-subscription-manager download-hysteria2
```

## Architecture

### Core Components
- **cmd/v2ray-manager/**: CLI application entry point
- **cmd/web-ui/**: Web UI server with REST API
- **internal/core/**: Core business logic
  - `parser/`: Subscription link parsing (VLESS, SS, Hysteria2)
  - `proxy/`: Proxy management (V2Ray, Hysteria2)
  - `workflow/`: Speed testing and auto-proxy workflows
  - `downloader/`: Binary downloaders for V2Ray and Hysteria2
- **web/static/**: Web UI frontend assets

### Web UI Architecture
The Web UI follows a clean architecture pattern:
- **Handlers**: HTTP request handlers (`cmd/web-ui/handlers/`)
- **Services**: Business logic layer (`cmd/web-ui/services/`)
- **Database**: SQLite with models (`cmd/web-ui/database/`)
- **Templates**: HTML templates with real-time updates

### Key Features
- **Multi-protocol support**: VLESS, Shadowsocks, Hysteria2
- **Batch speed testing**: High-concurrency testing (100+ threads)
- **Smart proxy management**: Auto-switching, port allocation
- **Port conflict detection**: Intelligent conflict resolution with user confirmation
- **Web UI**: Modern responsive interface with real-time updates
- **Dual-process architecture**: MVP tester + proxy server
- **Database state consistency**: Automatic cleanup and state synchronization

## Database

The project uses SQLite for persistent storage:
- **Location**: `data/v2ray_manager.db`
- **Tables**: subscriptions, nodes, proxy_status
- **Migrations**: Handled automatically on startup

## Important Files

### Configuration
- `configs/`: Configuration templates for V2Ray and Hysteria2
- `go.mod`: Go module dependencies

### Scripts
- `scripts/build.sh`: Multi-platform build script
- `scripts/build_webui.sh`: Web UI build script
- `start_web_ui.sh`: Quick Web UI starter
- `test_all_features.sh`: Integration test runner

### Runtime
- `v2ray/`: V2Ray core binaries and configurations
- `hysteria2/`: Hysteria2 client binaries
- `data/`: SQLite database and persistent data

## Development Notes

### Protocol Support
- **VLESS**: Full support via V2Ray core
- **Shadowsocks**: Auto-conversion for V2Ray 5.x compatibility
- **Hysteria2**: Uses separate Hysteria2 client
- **VMess/Trojan**: Planned for future versions

### Port Management
- Web UI: Default port 8888
- HTTP proxy: Default 7890 (configurable via settings)
- SOCKS proxy: Default 7891 (configurable via settings)
- Ports are dynamically allocated to avoid conflicts
- **Fixed port conflict handling**: System detects conflicts and prompts user confirmation before auto-disconnecting existing connections

### Error Handling
- Comprehensive error recovery and resource cleanup
- Graceful shutdown with signal handling
- Automatic process cleanup on startup

### Performance
- Supports 100+ concurrent speed tests
- Intelligent resource management
- Batch processing for large subscription lists
- Real-time progress reporting via SSE

## Common Development Patterns

### Adding New Protocol Support
1. Add parser in `internal/core/parser/protocols/`
2. Implement proxy logic in `internal/core/proxy/`
3. Update configuration templates in `configs/`
4. Add Web UI integration in services layer

### Database Changes
1. Update models in `cmd/web-ui/database/models.go`
2. Add migration logic in database initialization
3. Update service layer interfaces

### Web UI Extensions
1. Add API endpoints in `cmd/web-ui/handlers/`
2. Implement business logic in `cmd/web-ui/services/`
3. Update frontend JavaScript in `web/static/js/app.js`
4. Add CSS styling in `web/static/css/style.css`

## Critical Development Guidelines

### Database State Consistency ⚠️
**IMPORTANT**: Any operation that changes node connection status MUST maintain database consistency.

#### Required Pattern for Connection Operations
```go
// When stopping/disconnecting nodes:
1. Stop the proxy process
2. Update database node status to "idle"
3. Clear database port information (set to 0, 0)
4. Remove from memory connection map

// Example:
n.stopNodeConnection(connection)
n.setNodePorts(subscriptionID, nodeIndex, 0, 0)
n.updateNodeStatus(subscriptionID, nodeIndex, "idle")
delete(n.nodeConnections, key)
```

#### Functions That Must Update Database State
- `stopConnectionsByPort()` - Used in port conflict resolution
- `removeNodeConnection()` - Used in explicit disconnection
- `StopAllNodeConnections()` - Used in cleanup operations
- Any custom connection management functions

#### Startup State Cleanup
- System automatically cleans invalid states on startup via `cleanupNodeStatesOnStartup()`
- All "connected" or "connecting" states are reset to "idle" after restart
- Ensures database state matches actual running processes

#### Port Conflict Resolution Workflow
1. User attempts fixed port connection
2. System checks for conflicts via `/api/nodes/check-port-conflict`
3. If conflict detected, show detailed warning with conflicting node info
4. If user confirms, automatically stop conflicting connections AND update database
5. Start new connection and update database with new state

#### Testing Database Consistency
- Use `test_port_conflict_advanced.html` for comprehensive testing
- Always verify state persistence across page refreshes
- Check that "刷新页面模拟重启" shows consistent states

## API Reference

### Port Conflict Detection
**Endpoint**: `POST /api/nodes/check-port-conflict`

**Request**:
```json
{
  "port": 8888
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "has_conflict": true,
    "port": 8888,
    "protocol_type": "HTTP",
    "conflict_node_name": "节点名称",
    "conflict_node_index": 0,
    "subscription_id": "订阅ID"
  }
}
```

### Node Connection Operations
**Endpoint**: `POST /api/nodes/connect`

**Operations**:
- `http_fixed` - Connect to fixed HTTP port (with conflict detection)
- `socks_fixed` - Connect to fixed SOCKS port (with conflict detection)
- `http_random` - Connect to random HTTP port
- `socks_random` - Connect to random SOCKS port
- `disable` - Disconnect node and update database state

**Important**: All connection operations automatically handle database state updates.