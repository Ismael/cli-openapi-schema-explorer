# Project: openapi-explorer

A Go CLI tool for AI agents to progressively explore OpenAPI specs. Token-efficient YAML output with smart defaults.

## Architecture

```
cmd/           — Cobra commands (root, info, paths, components, tags)
internal/
  loader/      — Spec loading (file/URL, Swagger 2.0 auto-convert, OpenAPI 3.1 sanitization)
  resolver/    — $ref resolution (shallow default, full with --resolve)
  output/      — Formatting (FormatYAML, FormatList, FormatError)
testdata/      — Test specs (sample-api, complex-endpoint, paths-test, tags-test, many-paths, etc.)
```

## Key Design Decisions

- **Shallow $ref by default:** `Resolve()` keeps `$ref` in `responses.*.content.*.schema` and `requestBody.content.*.schema`. `ResolveFull()` inlines everything. The resolver tracks tree position via `effectiveKey()` to know when it's inside a body schema.
- **Auto-depth:** Path listing grows depth from 1 upward until output exceeds 50 lines. `--depth` overrides. `--filter` bypasses auto-depth entirely.
- **YAML only:** All structured output is YAML via `gopkg.in/yaml.v3`. No JSON output mode.
- **Config file:** `.openapi-explorer` YAML file in cwd provides `spec:` field. `--spec` flag overrides.
- **Cobra flag state in tests:** `resetCmdFlags()` in `cmd/info_test.go` resets all cobra flag `Changed` state and package-level variables between test runs. New flags/variables MUST be added there.

## Commands

- `info` — title, version, description, servers
- `paths [path] [method]` — progressive path exploration (3 modes: list, show path, show operation)
- `components [type] [name]` — progressive component exploration (3 modes: list types, list names, show detail)
- `tags` — list tag names with descriptions

## Flags

- `--spec` — spec file/URL (or use `.openapi-explorer` config)
- `--depth N` — override auto-depth (paths listing)
- `--filter` — filter paths by substring or tag (case-insensitive, paths listing)
- `--resolve` — full $ref inlining (paths operation detail)

## Testing

```bash
go test ./... -count=1
```

Test helpers are in `cmd/info_test.go`: `executeCommand()`, `resetCmdFlags()`, `testdataPath()`, `absTestdataPath()`.

All command tests use `executeCommand(args...)` which captures stdout via cobra's `SetOut`. Tests that parse structured output use `yaml.Unmarshal`.

## Adding a new flag

1. Add package-level variable in the relevant `cmd/*.go`
2. Register in `init()` with cobra
3. Add reset logic in `resetCmdFlags()` in `cmd/info_test.go`
4. Use `cmd.Flags().Changed("flagname")` to detect explicit usage (not the variable value)

## Spec compatibility

- OpenAPI 3.0.x: native support via kin-openapi
- OpenAPI 3.1: sanitization of numeric `exclusiveMinimum`/`exclusiveMaximum` to boolean
- Swagger 2.0: auto-conversion via `openapi2conv.ToV3()`
