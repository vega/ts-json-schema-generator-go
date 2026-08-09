// Programmatic wrapper around the native CLI: config in, parsed schema out.
// Keys mirror the ts-json-schema-generator 2.x Config type so that "config in →
// schema out" users can migrate by swapping createGenerator().createSchema()
// for generateSchema(). Dependency-free CommonJS.
"use strict";

const { spawn, spawnSync } = require("node:child_process");
const { resolveBinary } = require("./resolve-binary.js");

// spawnSync buffers the whole schema in memory and defaults to 1 MiB, which
// real projects exceed easily.
const MAX_BUFFER = 512 * 1024 * 1024;

// Each entry translates one config key into CLI arguments. Keys absent from
// the config, or set to undefined/null, contribute nothing.
const OPTIONS = {
    path: (value, args) => args.push("--path", string("path", value)),
    type: (value, args) => {
        for (const type of Array.isArray(value) ? value : [value]) {
            args.push("--type", string("type", type));
        }
    },
    tsconfig: (value, args) => args.push("--tsconfig", string("tsconfig", value)),
    schemaId: (value, args) => args.push("--id", string("schemaId", value)),
    expose: (value, args) => args.push("--expose", string("expose", value)),
    jsDoc: (value, args) => args.push("--jsDoc", string("jsDoc", value)),
    functions: (value, args) => args.push("--functions", string("functions", value)),
    markdownDescription: (value, args) => value && args.push("--markdown-description"),
    fullDescription: (value, args) => value && args.push("--full-description"),
    sortProps: (value, args) => value === false && args.push("--unstable"),
    topRef: (value, args) => value === false && args.push("--no-top-ref"),
    strictTuples: (value, args) => value === true && args.push("--strict-tuples"),
    skipTypeCheck: (value, args) => value === true && args.push("--no-type-check"),
    encodeRefs: (value, args) => value === false && args.push("--no-ref-encode"),
    additionalProperties: (value, args) => value === true && args.push("--additional-properties"),
    extraTags: (value, args) => {
        for (const tag of Array.isArray(value) ? value : [value]) {
            args.push("--validation-keywords", string("extraTags", tag));
        }
    },
    // The schema is returned as an object, so whitespace in the CLI's output
    // is immaterial.
    minify: () => {},
    discriminatorType: (value) => {
        if (value !== "json-schema") {
            throw new Error(
                `ts-json-schema-generator: discriminatorType ${JSON.stringify(value)} is not supported; ` +
                    `the native implementation always emits "json-schema" discriminators.`,
            );
        }
    },
};

function string(key, value) {
    if (typeof value !== "string") {
        throw new Error(`ts-json-schema-generator: ${key} must be a string, got ${typeof value}`);
    }
    return value;
}

function toArgs(config) {
    if (config === null || typeof config !== "object") {
        const got = config === null ? "null" : typeof config;
        throw new Error(`ts-json-schema-generator: config must be an object, got ${got}`);
    }

    const unsupported = Object.keys(config).filter((key) => !Object.hasOwn(OPTIONS, key));
    if (unsupported.length > 0) {
        throw new Error(
            `ts-json-schema-generator: unsupported config ${unsupported.length === 1 ? "option" : "options"} ` +
                `${unsupported.map((key) => JSON.stringify(key)).join(", ")}.\n` +
                `Supported options: ${Object.keys(OPTIONS).join(", ")}.\n` +
                `Because this package drives the native CLI in a separate process, options that carry ` +
                `JavaScript values cannot work: an existing \`tsProgram\`, and augmentors such as custom ` +
                `parsers, formatters, or SchemaGenerator subclasses. Stay on ts-json-schema-generator 2.x for those.`,
        );
    }

    const args = [];
    for (const [key, value] of Object.entries(config)) {
        if (value !== undefined && value !== null) {
            OPTIONS[key](value, args);
        }
    }
    return args;
}

function parseSchema(stdout) {
    try {
        return JSON.parse(stdout);
    } catch (error) {
        throw new Error(`ts-json-schema-generator: could not parse the generated schema: ${error.message}`);
    }
}

function failed(status, signal, stderr) {
    const reason = signal ? `was killed by ${signal}` : `exited with code ${status}`;
    const detail = stderr.trim();
    const error = new Error(`ts-json-schema-generator ${reason}${detail ? `:\n${detail}` : ""}`);
    error.exitCode = status;
    error.signal = signal ?? null;
    error.stderr = stderr;
    return error;
}

// generateSchema runs the native CLI and resolves with the parsed schema.
function generateSchema(config) {
    let args;
    let binary;
    try {
        args = toArgs(config);
        binary = resolveBinary();
    } catch (error) {
        return Promise.reject(error);
    }

    return new Promise((resolve, reject) => {
        const child = spawn(binary, args, { stdio: ["ignore", "pipe", "pipe"] });
        const stdout = [];
        const stderr = [];
        child.stdout.on("data", (chunk) => stdout.push(chunk));
        child.stderr.on("data", (chunk) => stderr.push(chunk));
        child.on("error", (error) => {
            reject(new Error(`ts-json-schema-generator: failed to run ${binary}: ${error.message}`));
        });
        child.on("close", (status, signal) => {
            if (status !== 0 || signal) {
                reject(failed(status, signal, Buffer.concat(stderr).toString()));
                return;
            }
            try {
                resolve(parseSchema(Buffer.concat(stdout)));
            } catch (error) {
                reject(error);
            }
        });
    });
}

// generateSchemaSync is generateSchema without the event loop.
function generateSchemaSync(config) {
    const args = toArgs(config);
    const binary = resolveBinary();

    const result = spawnSync(binary, args, { maxBuffer: MAX_BUFFER });
    if (result.error) {
        throw new Error(`ts-json-schema-generator: failed to run ${binary}: ${result.error.message}`);
    }
    if (result.status !== 0 || result.signal) {
        throw failed(result.status, result.signal, result.stderr.toString());
    }
    return parseSchema(result.stdout);
}

module.exports = { generateSchema, generateSchemaSync };
