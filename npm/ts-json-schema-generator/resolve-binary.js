// Locates the native CLI binary for the host. The binary itself lives in a
// per-platform optional dependency; TS_JSON_SCHEMA_GENERATOR_BINARY overrides
// the lookup with an explicit path. Dependency-free CommonJS so it runs before
// anything is built.
"use strict";

const SUPPORTED = "darwin-arm64, darwin-x64, linux-arm64, linux-x64, win32-arm64, win32-x64";

// resolveBinary returns the absolute path of the native CLI, or throws an
// Error whose message explains why no binary is available.
function resolveBinary() {
    const override = process.env.TS_JSON_SCHEMA_GENERATOR_BINARY;
    if (override) {
        return override;
    }

    const platform = process.platform;
    const arch = process.arch;
    const pkg = `ts-json-schema-generator-${platform}-${arch}`;
    const exe = platform === "win32" ? "bin/ts-json-schema-generator.exe" : "bin/ts-json-schema-generator";

    try {
        return require.resolve(`${pkg}/${exe}`);
    } catch {
        throw new Error(
            `ts-json-schema-generator: no native binary for ${platform}-${arch}.\n` +
                `The optional dependency "${pkg}" is not installed. It may not exist for this ` +
                `platform, or the install skipped optional dependencies (--no-optional, --omit=optional).\n` +
                `Supported platforms: ${SUPPORTED}.`,
        );
    }
}

module.exports = { resolveBinary };
