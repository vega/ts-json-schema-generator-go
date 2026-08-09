# ts-json-schema-generator

Generate JSON schema from your TypeScript sources.

Version 3 is a native binary built on [typescript-go](https://github.com/microsoft/typescript-go), the compiler that ships as TypeScript 7. The interface is the CLI; the package also exposes a thin [programmatic wrapper](#programmatic-usage) that runs it for you.

Installing this package pulls in a small platform-specific package containing the binary for your OS and CPU (macOS, Linux, and Windows on x64 and arm64).

## Install

```bash
npm install --save-dev ts-json-schema-generator@next
```

## Usage

```bash
npx ts-json-schema-generator --path 'my/project/**/*.ts' --type 'My.Type.Name'
```

The flags are the same as the 2.x CLI: `--path/-p`, `--type/-t`, `--tsconfig/-f`, `--id/-i`, `--expose/-e`, `--jsDoc/-j`, `--functions`, `--markdown-description`, `--full-description`, `--minify`, `--unstable`, `--strict-tuples`, `--no-top-ref`, `--no-type-check`, `--no-ref-encode`, `--additional-properties`, `--validation-keywords`, `--out/-o`. See [the options documentation](https://github.com/vega/ts-json-schema-generator#options) for what each one does.

## Programmatic usage

`generateSchema` takes the same options as the 2.x `Config` type and resolves with the parsed schema, so code that was "config in, schema out" migrates by swapping the call:

```js
import { generateSchema } from "ts-json-schema-generator";

const schema = await generateSchema({
    path: "src/types.ts",
    type: "MyType",
    tsconfig: "tsconfig.json",
});
```

`generateSchemaSync` is the same thing without the `await`:

```js
import { generateSchemaSync } from "ts-json-schema-generator";

const schema = generateSchemaSync({ path: "src/types.ts", type: "MyType" });
```

Both spawn the native binary and throw (or reject) with an `Error` carrying the CLI's `stderr` when generation fails. TypeScript definitions ship with the package.

Options that carry JavaScript values cannot cross the process boundary: an existing `tsProgram`, and augmentors such as custom node parsers, type formatters, or `SchemaGenerator` subclasses. Those are unsupported here, and passing them throws — projects that need them should stay on the 2.x line, which remains the Node.js library.

Advanced: set `TS_JSON_SCHEMA_GENERATOR_BINARY` to run a specific binary instead of the one from the platform package. It applies to the CLI too, and is mainly useful for testing a local build.

## Links

- Source: [vega/ts-json-schema-generator-go](https://github.com/vega/ts-json-schema-generator-go)
- The TypeScript implementation (2.x, library + CLI): [vega/ts-json-schema-generator](https://github.com/vega/ts-json-schema-generator)

MIT licensed.
