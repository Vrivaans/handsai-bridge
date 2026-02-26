# HandsAI Bridge (Go)

A lightweight MCP (Model Context Protocol) server written in Go that acts as a bridge between any MCP-compatible IDE client and the [HandsAI v3](https://github.com/Vrivaans/handsaiv3) Spring Boot backend.

It translates **JSON-RPC over stdio** (the MCP standard) into plain **HTTP REST calls** to the HandsAI API, and back.

> The previous Node.js/TypeScript implementation has been deprecated in favor of this Go version due to its universal compatibility with restricted IDE environments (like Antigravity and Claude Desktop) where runtime dependencies like `node`, `npx`, or `tsx` are often unavailable.

## Why Go?

- **Zero runtime dependencies** — compile once, run anywhere. No Node, no npm, no PATH issues.
- **Single binary** — one executable file, easy to distribute.
- **IDE-agnostic** — works identically in Antigravity, Claude Desktop, VS Code, and any other MCP client.
- **Fast startup** — no JVM warmup, no package loading.

## Prerequisites

- [Go 1.21+](https://golang.org/dl/) — only needed to **build** from source.
- A running instance of [HandsAI v3](https://github.com/Vrivaans/handsaiv3) (default: `http://localhost:8080`).

## Quick Start

### Option A: Use the pre-compiled binary

A pre-compiled `handsai-mcp` binary for macOS (darwin/arm64) is included in this repo. Just make it executable:

```bash
chmod +x handsai-mcp
```

### Option B: Build from source

```bash
git clone https://github.com/Vrivaans/handsai-bridge.git
cd handsai-bridge
go build -o handsai-mcp main.go
```

## Configuration

By default, the bridge connects to `http://localhost:8080`.

To change the port or host, create a `config.json` file **in the same directory as the binary**:

```json
{
  "handsaiUrl": "http://localhost:9090"
}
```

If the file doesn't exist, the default is used automatically.

## IDE Integration

### Antigravity / Claude Desktop / Any MCP Client

Add the following to your `mcp_config.json` (Antigravity) or `claude_desktop_config.json` (Claude Desktop):

```json
{
  "mcpServers": {
    "handsai": {
      "command": "/absolute/path/to/handsai-mcp",
      "args": ["mcp"]
    }
  }
}
```

> **Important:** Use the **absolute path** to the binary. The `args: ["mcp"]` field is required by some IDE clients to properly register the server.

## How It Works

```
IDE (MCP Client)  →  stdio JSON-RPC  →  handsai-mcp (Go)  →  HTTP  →  HandsAI (Spring Boot)
```

1. The IDE spawns `handsai-mcp` as a subprocess.
2. The bridge reads JSON-RPC messages from `stdin` line by line.
3. For `tools/list`, it calls `GET /mcp/tools/list` on HandsAI.
4. For `tools/call`, it calls `POST /mcp/tools/call` on HandsAI.
5. Responses are written back to `stdout` as JSON-RPC.

## Cross-Compilation

Build for other platforms from macOS:

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o handsai-mcp-linux main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o handsai-mcp.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o handsai-mcp-intel main.go
```

## Related Projects

- [HandsAI v3](https://github.com/Vrivaans/handsaiv3) — The Spring Boot backend this bridge connects to.
