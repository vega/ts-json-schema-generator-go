#!/usr/bin/env node
// Launcher for the native CLI. The binary itself lives in a per-platform
// optional dependency; this script locates the one matching the host and
// execs it. Dependency-free CommonJS so it runs before anything is built.
"use strict";

const { execFileSync } = require("node:child_process");

const platform = process.platform;
const arch = process.arch;
const pkg = `ts-json-schema-generator-${platform}-${arch}`;
const exe = platform === "win32" ? "bin/ts-json-schema-generator.exe" : "bin/ts-json-schema-generator";

let binary;
try {
    binary = require.resolve(`${pkg}/${exe}`);
} catch {
    console.error(
        `ts-json-schema-generator: no native binary for ${platform}-${arch}.\n` +
            `The optional dependency "${pkg}" is not installed. It may not exist for this ` +
            `platform, or the install skipped optional dependencies (--no-optional, --omit=optional).\n` +
            `Supported platforms: darwin-arm64, darwin-x64, linux-arm64, linux-x64, win32-arm64, win32-x64.`,
    );
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
