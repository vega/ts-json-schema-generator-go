# Agent guide

This codebase was **fully AI-generated**: it is a port of
[vega/ts-json-schema-generator](https://github.com/vega/ts-json-schema-generator)
from TypeScript to Go, written by Claude (Anthropic) in August 2026 using
parallel subagents, and reviewed/merged by the same. Treat it accordingly —
prefer verifying behavior against the reference implementation and the test
oracle over trusting comments or structure.

## The reference implementation is the spec

The TypeScript sources in vega/ts-json-schema-generator (`src/`, `factory/`)
are the behavioral reference. The Go code mirrors them file-by-file:

| Go | TypeScript |
|---|---|
| `internal/types` | `src/Type/` |
| `internal/parser` | `src/NodeParser/`, `src/AnnotationsReader/` |
| `internal/formatter` | `src/TypeFormatter/` |
| `internal/generator` | `src/SchemaGenerator.ts` |
| `internal/factory` | `factory/` |
| `cmd/ts-json-schema-generator` | `ts-json-schema-generator.ts` |

When fixing a behavior bug, read the corresponding TS file first and port its
semantics faithfully — including odd edge cases. Do not "improve" behavior in
ways that change generated schemas.

## Test oracle

```bash
npm install        # once; vega-lite sources for the regression test
go test ./...      # 251 golden-schema fixtures + full vega-lite schema, ~10 s
```

- `test/valid-data/<name>/` holds each fixture (`main.ts` + golden
  `schema.json`); `test/fixtures-manifest.json` records generator options per
  fixture (regenerate with `go run ./tools/extract_fixtures` if you edit a
  fixture's `index.test.ts`).
- Schemas are compared as parsed JSON (order-insensitive). If output differs
  from a golden file, the golden may only be updated when the new output is
  *semantically equivalent* (e.g. same enum value set in a different
  deterministic order) or the golden is demonstrably stale — say so in the
  commit message.
- The e2e tests chdir to the repo root (node-key hashes embed cwd-relative
  file names, mirroring `process.cwd()` in the original).

## typescript-go and the shim layer

The TypeScript 7 compiler (microsoft/typescript-go) keeps its API in
`internal/` packages. This repo imports them through generated shim modules in
`shim/` whose module paths are declared under
`github.com/microsoft/typescript-go/shim/...`, which makes the internal
imports legal. The compiler comes from the Go module proxy at a pinned
pseudo-version — no fork, no submodule, no patches.

- Need an unexported checker method/field? Add it to
  `shim/checker/extra-shim.json` and run `go run ./tools/gen_shims`; it
  appears as `checker.Checker_<name>(recv, ...)`.
- Bumping the compiler: update the pinned version in every `shim/*/go.mod`,
  `go mod tidy` in each shim dir and the root, rerun `gen_shims`, then run the
  full test suite.
- Because the shims are wired with `replace` directives, this module is not
  `go install`-able from the proxy; build from a clone.

## Conventions and gotchas

- **Errors**: parsers and formatters `panic(error)` where the TS code throws;
  `generator.CreateSchema` recovers at the boundary. Set `TSJSG_DEBUG_PANIC=1`
  to disable the recover and get stack traces.
- **Numbers** are `float64`; format with `types.NumberToString` (JS
  `Number#toString` semantics). Literal values are `string | float64 | bool`
  in `any`-typed fields.
- **tsgo API differences** already handled — keep using the helpers:
  - `tsutils.GetSymbolAtLocation` (the raw checker method panics on
    synthesized nodes);
  - `TypeToTypeNode` records synthesized identifier symbols in the
    `synthesizedSymbols` registry (`internal/parser/node_text.go`) — pass it
    at every call site;
  - `tsutils.HasJSDocTag` has a text fallback for `@@tag` comments (the old
    parser produced a real tag from `@@hidden`; vega-lite depends on it).
- **Ordering matters**: parser/formatter registration order in
  `internal/factory/wiring.go` is load-bearing (first `SupportsType`/
  `SupportsNode` match wins) and mirrors `factory/parser.ts` /
  `factory/formatter.ts` exactly. Keep it in sync with the reference.
- The generator is **not safe for concurrent use** (package-level caches,
  cwd-dependent node keys). One generation at a time per process.

## Known divergences from the TypeScript version

Deliberate and documented in commit history:

1. `keyof`-derived literal unions enumerate in the native checker's order
   (deterministic, same value sets; affected goldens were updated).
2. JSDoc `{@link}` text rendering matches the services-layer display parts.
3. `keyof` also honors index signatures inherited from base interfaces
   (the TypeScript implementation only checks own signatures; TypeScript 7's
   lib files moved several onto base interfaces, e.g. CSSStyleDeclaration).
4. Sourceless (synthesized) node keys use the node address instead of
   `Math.random()` — stable within a run, so caching works.
5. The programmatic Node.js API is not ported; the CLI is the interface. The
   npm package wraps it with `generateSchema`/`generateSchemaSync` (config in,
   schema out), which cover config-only usage — `tsProgram` and the augmentors
   have no equivalent across a process boundary.

## Downstream regression tests

- `internal/e2e/vegalite_test.go` — full vega-lite schema (sources from npm).
- `internal/e2e/mosaic_test.go` — Mosaic/vgplot spec schema from the vendored
  `test/mosaic` snapshot, using mosaic's own CLI invocation; CSSStyles
  properties are compared as a superset because they track the lib.dom
  version.

## Development workflow

Changes go through issues and pull requests: file (or find) an issue, work
on a branch, open a PR whose description says `Closes #N`, and leave merging
to a human reviewer. Do not commit to `main` directly.

## Keeping up with typescript-go

`tools/bump-tsgo.sh <ref>` re-pins the compiler and regenerates shims; the
weekly `bump-typescript-go` workflow does this against `main` and opens a PR
only if the full test oracle passes.
