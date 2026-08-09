// Fetches @uwdata/mosaic-spec into node_modules for the mosaic regression
// test. Installed via `npm pack` + extraction rather than as a regular
// dependency because the package's transitive dependencies do not install
// under npm (workspace: protocol); the spec sources themselves are
// self-contained and the tarball includes the published schema.
import { execSync } from "node:child_process";
import { existsSync, mkdirSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const version = "0.29.2";
const target = join("node_modules", "@uwdata", "mosaic-spec");

if (existsSync(join(target, "src", "spec", "Spec.ts"))) {
    process.exit(0);
}

const work = join(tmpdir(), `mosaic-spec-${process.pid}`);
mkdirSync(work, { recursive: true });
try {
    const tarball = execSync(`npm pack @uwdata/mosaic-spec@${version} --silent`, { cwd: work })
        .toString()
        .trim();
    mkdirSync(target, { recursive: true });
    execSync(`tar xzf ${join(work, tarball)} -C ${target} --strip-components=1`);
    console.log(`fetched @uwdata/mosaic-spec@${version}`);
} finally {
    rmSync(work, { recursive: true, force: true });
}
