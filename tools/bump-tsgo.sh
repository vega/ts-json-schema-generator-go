#!/usr/bin/env bash
# Bump the pinned typescript-go version and regenerate the shim layer.
#
# Usage: tools/bump-tsgo.sh [ref]
#   ref: a typescript-go branch, tag, or commit
#        (default: the latest TypeScript release tag, typescript/v*)
#
# The fixture suite is the compatibility oracle — run `go test ./...` after.
set -euo pipefail
cd "$(dirname "$0")/.."

# The upstream compiler module. Override to track a fork, a repo move, or a
# major-version path (…/v2).
TSGO_MODULE="${TSGO_MODULE:-github.com/microsoft/typescript-go}"
# Module paths with a major-version suffix don't name a git repo.
tsgo_repo="https://${TSGO_MODULE%/v[0-9]*}"

ref="${1:-}"
if [ -z "$ref" ]; then
    # Module queries reject refs containing slashes, so resolve the newest
    # typescript/v* release tag to its commit (preferring the peeled entry
    # of the annotated tag).
    tags=$(git ls-remote --tags "$tsgo_repo" 'typescript/v*')
    tag=$(printf '%s\n' "$tags" | sed 's|.*refs/tags/||; s|\^{}$||' | sort -u -V | tail -1)
    ref=$(printf '%s\n' "$tags" | grep -F "refs/tags/${tag}^{}" | cut -f1 || true)
    if [ -z "$ref" ]; then
        ref=$(printf '%s\n' "$tags" | grep -F "refs/tags/${tag}" | head -1 | cut -f1 || true)
    fi
    if [ -z "$ref" ]; then
        echo "could not resolve a typescript/v* release tag" >&2
        exit 1
    fi
    echo "Latest release: ${tag} (${ref})"
fi

echo "Resolving ${TSGO_MODULE}@${ref}..."
version=$(go list -m -json "${TSGO_MODULE}@${ref}" | go run ./tools/internal/jsonfield Version)
echo "Pinning ${version}"

for mod in shim/*/go.mod shim/vfs/*/go.mod; do
    [ -f "$mod" ] || continue
    dir=$(dirname "$mod")
    (cd "$dir" && go mod edit -require="${TSGO_MODULE}@${version}" && go mod tidy >/dev/null)
done

# tidy alone never downgrades the root requirement, so pin it explicitly.
go mod edit -require="${TSGO_MODULE}@${version}"
go mod tidy
echo "Regenerating shims..."
go run ./tools/gen_shims
go mod tidy

echo "Done. Now run: go test ./..."
