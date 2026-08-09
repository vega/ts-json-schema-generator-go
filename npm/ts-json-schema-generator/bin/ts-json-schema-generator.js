#!/usr/bin/env node
// Launcher for the native CLI: locates the binary matching the host and execs
// it. Dependency-free CommonJS so it runs before anything is built.
"use strict";

const { execFileSync } = require("node:child_process");
const { resolveBinary } = require("../resolve-binary.js");

let binary;
try {
    binary = resolveBinary();
} catch (error) {
    console.error(error.message);
    process.exit(1);
}

try {
    execFileSync(binary, process.argv.slice(2), { stdio: "inherit" });
} catch (error) {
    if (typeof error.status === "number") {
        process.exit(error.status);
    }
    console.error(`ts-json-schema-generator: failed to run ${binary}: ${error.message}`);
    process.exit(1);
}
