// Assembles the npm distribution into dist-npm/: one package per platform
// carrying a single prebuilt binary, plus the main package (copied from
// npm/ts-json-schema-generator) whose version and optionalDependencies are
// stamped with the release version.
//
// Usage:
//   node tools/make-npm-packages.mjs --version 3.0.0-next.0 --binaries <dir>
//
// <dir> holds the binaries as named by the release workflow's build job:
// ts-json-schema-generator-<goos>-<goarch>[.exe].
import { chmodSync, copyFileSync, existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const TARGETS = [
    { goos: "darwin", goarch: "arm64", os: "darwin", cpu: "arm64" },
    { goos: "darwin", goarch: "amd64", os: "darwin", cpu: "x64" },
    { goos: "linux", goarch: "arm64", os: "linux", cpu: "arm64" },
    { goos: "linux", goarch: "amd64", os: "linux", cpu: "x64" },
    { goos: "windows", goarch: "arm64", os: "win32", cpu: "arm64" },
    { goos: "windows", goarch: "amd64", os: "win32", cpu: "x64" },
];

const REPO_URL = "git+https://github.com/vega/ts-json-schema-generator-go.git";

function parseArgs(argv) {
    const args = { version: "", binaries: "" };
    for (let i = 0; i < argv.length; i += 2) {
        const flag = argv[i];
        const value = argv[i + 1];
        if (flag === "--version") {
            args.version = value;
        } else if (flag === "--binaries") {
            args.binaries = value;
        } else {
            throw new Error(`unknown argument: ${flag}`);
        }
        if (value === undefined) {
            throw new Error(`${flag} requires a value`);
        }
    }
    if (!args.version) {
        throw new Error("--version is required");
    }
    if (!args.binaries) {
        throw new Error("--binaries is required");
    }
    return args;
}

const { version, binaries } = parseArgs(process.argv.slice(2));

const root = new URL("..", import.meta.url).pathname;
const source = join(root, "npm", "ts-json-schema-generator");
const dist = join(root, "dist-npm");

rmSync(dist, { recursive: true, force: true });
mkdirSync(dist, { recursive: true });

const built = [];
for (const target of TARGETS) {
    const ext = target.goos === "windows" ? ".exe" : "";
    const binary = join(binaries, `ts-json-schema-generator-${target.goos}-${target.goarch}${ext}`);
    if (!existsSync(binary)) {
        console.warn(`warning: skipping ${target.os}-${target.cpu}, no binary at ${binary}`);
        continue;
    }

    const name = `ts-json-schema-generator-${target.os}-${target.cpu}`;
    const dir = join(dist, name);
    mkdirSync(join(dir, "bin"), { recursive: true });

    const destination = join(dir, "bin", `ts-json-schema-generator${ext}`);
    copyFileSync(binary, destination);
    chmodSync(destination, 0o755);

    writeFileSync(
        join(dir, "package.json"),
        `${JSON.stringify(
            {
                name,
                version,
                description: `The ${target.os}-${target.cpu} binary for ts-json-schema-generator`,
                license: "MIT",
                repository: { type: "git", url: REPO_URL },
                os: [target.os],
                cpu: [target.cpu],
                engines: { node: ">=20" },
            },
            null,
            2,
        )}\n`,
    );

    copyFileSync(join(root, "LICENSE"), join(dir, "LICENSE"));

    built.push(name);
    console.log(`packaged ${name}@${version}`);
}

const manifest = JSON.parse(readFileSync(join(source, "package.json"), "utf8"));
manifest.version = version;
for (const dependency of Object.keys(manifest.optionalDependencies)) {
    manifest.optionalDependencies[dependency] = version;
}

const mainDir = join(dist, "ts-json-schema-generator");
mkdirSync(join(mainDir, "bin"), { recursive: true });
writeFileSync(join(mainDir, "package.json"), `${JSON.stringify(manifest, null, 2)}\n`);
copyFileSync(join(source, "bin", "ts-json-schema-generator.js"), join(mainDir, "bin", "ts-json-schema-generator.js"));
chmodSync(join(mainDir, "bin", "ts-json-schema-generator.js"), 0o755);
for (const file of ["index.js", "index.d.ts", "resolve-binary.js", "README.md"]) {
    copyFileSync(join(source, file), join(mainDir, file));
}
copyFileSync(join(root, "LICENSE"), join(mainDir, "LICENSE"));
console.log(`packaged ts-json-schema-generator@${version} (${built.length}/${TARGETS.length} platforms)`);
