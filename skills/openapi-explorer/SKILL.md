---
name: openapi-explorer
description: Token-efficient CLI for exploring OpenAPI specs with batch query support. Use when the user wants to explore, understand, or query an OpenAPI/Swagger API specification.
---

# openapi-explorer

CLI tool for exploring OpenAPI specifications.

## Discovery Flow

There are only 4 commands: `info`, `paths`, `components`, `tags`.

1. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer paths` — list paths with methods (auto-depth)
2. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer paths --filter users` — filter paths by substring or tag name
3. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer paths /users` — list sub-paths and methods
3. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer paths /users get` — operation detail ($ref kept for body schemas)
4. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer paths /users get post` — multiple methods at once
5. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer paths /users get --resolve` — operation detail (all $ref inlined)

For components:
1. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer components` — list component types
2. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer components schemas` — list schema names
3. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer components schemas User` — full schema detail (always resolved)
4. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer components schemas User Order` — multiple schemas at once

For tags:
1. `${CLAUDE_PLUGIN_ROOT}/bin/openapi-explorer tags` — list tags with descriptions

## Commands

Only these 4 commands exist:
- `info` — spec metadata (title, version, description, servers)
- `paths [path] [methods]` — explore paths (path + methods gives operations detail)
- `components [type] [names]` — explore components
- `tags` — list API tags with descriptions

## Flags

- `--spec <file-or-url>` — OpenAPI spec (or use `.openapi-explorer` config)
- `--depth N` — path listing depth (default: auto, stays under 50 lines)
- `--filter <text>` — filter paths by substring or tag name (case-insensitive)
- `--resolve` — fully resolve all $ref inline (paths command only)

