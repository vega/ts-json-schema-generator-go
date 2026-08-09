# Mosaic spec fixture

Vendored snapshot of the `src/spec` TypeScript sources and the published
`dist/mosaic-schema.json` from [`@uwdata/mosaic-spec`](https://github.com/uwdata/mosaic)
v0.29.2 (BSD-3-Clause, see LICENSE). Mosaic generates its schema with:

    ts-json-schema-generator -f tsconfig.json -p src/spec/Spec.ts -t Spec \
        --no-type-check --no-ref-encode --functions hide

The regression test in internal/e2e replays that invocation with this port
and compares against the published schema. Vendored (rather than installed
from npm like vega-lite) because the mosaic packages' transitive
dependencies currently fail to install under npm.
