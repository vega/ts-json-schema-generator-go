// Exercises the programmatic API against a real build of the native CLI.
// Set TS_JSON_SCHEMA_GENERATOR_BINARY to an existing binary to skip the build.
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { after, before, describe, it } from "node:test";
import { fileURLToPath } from "node:url";

const { generateSchema, generateSchemaSync } = await import("./index.js");

const repo = fileURLToPath(new URL("../../", import.meta.url));

let fixture;
let tsconfig;

before(
    () => {
        if (!process.env.TS_JSON_SCHEMA_GENERATOR_BINARY) {
            const build = join(mkdtempSync(join(tmpdir(), "tsjsg-bin-")), "bin");
            const name = process.platform === "win32" ? "ts-json-schema-generator.exe" : "ts-json-schema-generator";
            const exe = join(build, name);
            execFileSync("go", ["build", "-o", exe, "./cmd/ts-json-schema-generator"], { cwd: repo, stdio: "inherit" });
            process.env.TS_JSON_SCHEMA_GENERATOR_BINARY = exe;
        }

        const dir = mkdtempSync(join(tmpdir(), "tsjsg-fixture-"));
        fixture = join(dir, "main.ts");
        tsconfig = join(dir, "tsconfig.json");
        writeFileSync(
            fixture,
            `export interface MyObject {
    /**
     * The name of the thing.
     */
    name: string;
    count?: number;
    kind: "a" | "b";
}

export interface OtherObject {
    id: string;
}
`,
        );
        writeFileSync(tsconfig, `{ "compilerOptions": { "strict": true } }\n`);
    },
    { timeout: 600_000 },
);

after(() => {
    delete process.env.TS_JSON_SCHEMA_GENERATOR_BINARY;
});

function config(overrides) {
    return { path: fixture, tsconfig, type: "MyObject", ...overrides };
}

describe("generateSchema", () => {
    it("generates a schema from a config", async () => {
        const schema = await generateSchema(config());

        assert.equal(schema.$schema, "http://json-schema.org/draft-07/schema#");
        assert.equal(schema.$ref, "#/definitions/MyObject");
        assert.deepEqual(Object.keys(schema.definitions), ["MyObject"]);

        const object = schema.definitions.MyObject;
        assert.equal(object.type, "object");
        assert.equal(object.additionalProperties, false);
        assert.equal(object.properties.name.type, "string");
        assert.equal(object.properties.name.description, "The name of the thing.");
        assert.deepEqual(object.properties.kind.enum, ["a", "b"]);
        assert.deepEqual(object.required, ["name", "kind"]);
    });

    it("maps the option flags", async () => {
        const schema = await generateSchema(
            config({
                schemaId: "https://example.com/my.json",
                topRef: false,
                sortProps: false,
                additionalProperties: true,
                minify: true,
            }),
        );

        // Without a top-level $ref the root type is inlined.
        assert.equal(schema.$id, "https://example.com/my.json");
        assert.equal(schema.$ref, undefined);
        // Allowing additional properties drops the `additionalProperties: false`
        // the first test sees.
        assert.equal(schema.additionalProperties, undefined);
        // Sorted output would be count, kind, name.
        assert.deepEqual(Object.keys(schema.properties), ["name", "count", "kind"]);
    });

    it("accepts several types", async () => {
        const schema = await generateSchema(config({ type: ["MyObject", "OtherObject"] }));

        assert.deepEqual(Object.keys(schema.definitions).sort(), ["MyObject", "OtherObject"]);
    });

    it("rejects with the CLI's stderr", async () => {
        const error = await generateSchema(config({ type: "NoSuchType" })).then(
            () => assert.fail("expected a rejection"),
            (error) => error,
        );

        assert.match(error.message, /NoSuchType/);
        assert.match(error.stderr, /NoSuchType/);
        assert.equal(error.exitCode, 1);
    });

    it("rejects unsupported config keys", async () => {
        const error = await generateSchema(config({ tsProgram: {}, augmentors: [] })).then(
            () => assert.fail("expected a rejection"),
            (error) => error,
        );

        assert.match(error.message, /"tsProgram", "augmentors"/);
        assert.match(error.message, /2\.x/);
    });
});

describe("generateSchemaSync", () => {
    it("generates the same schema", () => {
        const schema = generateSchemaSync(config());

        assert.equal(schema.$ref, "#/definitions/MyObject");
        assert.equal(schema.definitions.MyObject.properties.count.type, "number");
    });

    it("throws with the CLI's stderr", () => {
        assert.throws(() => generateSchemaSync(config({ type: "NoSuchType" })), {
            message: /NoSuchType/,
            exitCode: 1,
        });
    });

    it("throws on a non-string option", () => {
        assert.throws(() => generateSchemaSync(config({ expose: true })), { message: /expose must be a string/ });
    });
});
