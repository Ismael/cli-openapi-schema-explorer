# openapi-explorer

A CLI tool for progressive, token-efficient exploration of OpenAPI specifications. Designed for AI agent consumption.

Supports OpenAPI 3.x and Swagger 2.0 specs from local files or URLs.

## Install

### Claude Code Plugin

Inside Claude Code, run:

```
# Add the marketplace (one-time)
/plugin marketplace add Ismael/cli-openapi-schema-explorer

# Install the plugin
/plugin install openapi-explorer@openapi-explorer
```

The install script automatically downloads the correct binary for your platform.

### From source

```bash
go install github.com/Ismael/cli-openapi-schema-explorer@latest
```

### From GitHub releases

Download the binary for your platform from the [releases page](https://github.com/Ismael/cli-openapi-schema-explorer/releases/latest).

## Quick Start

```bash
# Set up config (optional — or use --spec on every call)
echo "spec: api.yaml" > .openapi-explorer

# List API paths with methods (auto-depth)
openapi-explorer paths

# Filter paths by keyword or tag
openapi-explorer paths --filter clip

# List tags with descriptions
openapi-explorer tags

# Explore a specific path
openapi-explorer paths /users

# Get operation detail (body schemas as $ref)
openapi-explorer paths /users get

# Get operation detail with full resolution
openapi-explorer paths /users get --resolve

# Explore components
openapi-explorer components schemas User
```

See [SKILL.md](skills/openapi-explorer/SKILL.md) for the full agent usage guide.

## Credits

Based on [kadykov/mcp-openapi-schema-explorer](https://github.com/kadykov/mcp-openapi-schema-explorer) — an MCP server that provides token-efficient access to OpenAPI specifications via MCP Resource Templates.
