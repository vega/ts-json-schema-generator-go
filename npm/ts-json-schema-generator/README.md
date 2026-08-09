# ts-json-schema-generator

Generate JSON schema from your TypeScript sources.

Version 3 is a native binary built on [typescript-go](https://github.com/microsoft/typescript-go), the compiler that ships as TypeScript 7. It is **CLI-only** — there is no programmatic Node.js API. Projects that import `ts-json-schema-generator` as a library should stay on the 2.x line.

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

## Links

- Source: [vega/ts-json-schema-generator-go](https://github.com/vega/ts-json-schema-generator-go)
- The TypeScript implementation (2.x, library + CLI): [vega/ts-json-schema-generator](https://github.com/vega/ts-json-schema-generator)

MIT licensed.
