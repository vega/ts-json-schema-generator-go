# ts-json-schema-generator-go

Generate JSON schema from your TypeScript sources — implemented in Go on top of [typescript-go](https://github.com/microsoft/typescript-go), the native compiler released as TypeScript 7.

This is the native port of [vega/ts-json-schema-generator](https://github.com/vega/ts-json-schema-generator). The TypeScript implementation over there is the reference: the Go code mirrors it module for module, and both produce the same schemas (verified by 251 golden-schema fixtures plus full [vega-lite](https://github.com/vega/vega-lite) and [Mosaic](https://github.com/uwdata/mosaic) schema comparisons). The CLI is a single binary with no Node.js dependency, and the whole test corpus runs in seconds.

> [!NOTE]
> This port was fully AI-generated (by Claude, against the TypeScript reference implementation and its golden-schema test suite). See [AGENTS.md](AGENTS.md) for provenance and working conventions.

## Install from npm

```bash
npm install --save-dev ts-json-schema-generator@native
npx ts-json-schema-generator --path 'my/project/**/*.ts' --type 'My.Type.Name'
```

The `native` dist-tag of [`ts-json-schema-generator`](https://www.npmjs.com/package/ts-json-schema-generator) is this CLI — the package pulls in a platform-specific binary via one of the platform packages; `2.x` remains the Node.js library, and the `next`/`canary` tags belong to its release automation.

The package also exports `generateSchema(config)` and `generateSchemaSync(config)`, which run the binary and return the parsed schema:

```js
const schema = await generateSchema({ path: "src/types.ts", type: "MyType" });
```

The config keys mirror the 2.x `Config` type. Anything that carries a JavaScript value across the process boundary — an existing `tsProgram`, custom parsers, formatters, or `SchemaGenerator` subclasses — is unsupported and throws; those uses belong on 2.x.

Platform binary packages (installed automatically as optional dependencies):
[darwin-arm64](https://www.npmjs.com/package/ts-json-schema-generator-darwin-arm64) ·
[darwin-x64](https://www.npmjs.com/package/ts-json-schema-generator-darwin-x64) ·
[linux-x64](https://www.npmjs.com/package/ts-json-schema-generator-linux-x64) ·
[linux-arm64](https://www.npmjs.com/package/ts-json-schema-generator-linux-arm64) ·
[win32-x64](https://www.npmjs.com/package/ts-json-schema-generator-win32-x64) ·
[win32-arm64](https://www.npmjs.com/package/ts-json-schema-generator-win32-arm64)

The packaging sources live in [`npm/`](npm) and are assembled by [`tools/make-npm-packages.mjs`](tools/make-npm-packages.mjs); releases publish via npm trusted publishing from [`release.yml`](.github/workflows/release.yml).

## Usage

```bash
ts-json-schema-generator --path 'my/project/**/*.ts' --type 'My.Type.Name'
```

The flags mirror the original CLI: `--path/-p`, `--type/-t`, `--tsconfig/-f`, `--id/-i`, `--expose/-e`, `--jsDoc/-j`, `--functions`, `--markdown-description`, `--full-description`, `--minify`, `--unstable`, `--no-top-ref`, `--no-type-check`, `--no-ref-encode`, `--additional-properties`, `--validation-keywords`, `--out/-o`. See [the original documentation](https://github.com/vega/ts-json-schema-generator#options) for what they do.

### `--outdir`: one file per type

`--outdir <dir>` is the one flag this port adds on top of the original CLI. It writes a separate schema file per requested type to `<dir>/<type>.schema.json`:

```bash
ts-json-schema-generator --path 'src/**/*.ts' --outdir schemas --type Spec --type Config
# schemas/Spec.schema.json, schemas/Config.schema.json
```

Each file is exactly the schema that a single `--type` run would produce (a top-level `$ref` plus that type's reachable definitions), and every other flag applies to all of them. `--outdir` cannot be combined with `--out`, needs the types spelled out (no `*`, no duplicates), and rejects type names that would not make sound file names.

What it saves is the work that happens once per process: loading the program, parsing, and — unless you pass `--no-type-check` — checking the whole project. Only the per-type schema walk repeats. Generating eight vega-lite types this way takes about half as long as eight separate invocations with `--no-type-check`; leave type checking on, as it is by default, and the gap widens, because the full-project check is the part being amortized. For a single type there is nothing to amortize and `--out` is equivalent.

Two things to know before wiring it into a build. Files are written as each type completes, so a failure partway through — an unknown type name, say — leaves the files generated before it on disk, and the error names the type it failed on. And `--id` sets the same `$id` on every file, since each file is just what its single-type run would have produced; that will collide in validators and resolvers that cache schemas by `$id`, so either leave `--id` off or set the per-file `$id` yourself afterwards.

## Building

```bash
go build -o ts-json-schema-generator ./cmd/ts-json-schema-generator
```

To run the full test suite (the vega-lite regression test needs the vega-lite sources from npm):

```bash
npm install
go test ./...
```

## How it uses typescript-go

typescript-go keeps its compiler API in `internal/` packages. This repo accesses them through generated shim modules under `shim/` whose module paths are declared beneath `github.com/microsoft/typescript-go/shim/...`, which makes the internal imports legal; the compiler itself is pulled from the Go module proxy at a pinned version (no fork, no submodule, no patches). The shims are generated by `tools/gen_shims` — bump the pinned version in the `shim/*/go.mod` files and rerun it to upgrade the compiler.

Because the shims are wired up with `replace` directives, this module cannot be consumed with `go install`/`go get` from the proxy; clone and build, or use a released binary.

## Architecture

| Go package | Mirrors (in [ts-json-schema-generator](https://github.com/vega/ts-json-schema-generator)) |
|---|---|
| `internal/types` | `src/Type/` — the intermediate type model |
| `internal/parser` | `src/NodeParser/`, `src/AnnotationsReader/` — AST → type model |
| `internal/formatter` | `src/TypeFormatter/` — type model → JSON schema |
| `internal/generator` | `src/SchemaGenerator.ts` |
| `internal/factory` | `factory/` — program creation and chain wiring |
| `cmd/ts-json-schema-generator` | `ts-json-schema-generator.ts` — the CLI |

Test fixtures in `test/valid-data` are shared with the original project; `test/fixtures-manifest.json` (generated by `tools/extract_fixtures`) records each fixture's generator options, and `internal/e2e` replays them all plus the vega-lite golden schema.

## License

MIT — see [LICENSE](LICENSE).
