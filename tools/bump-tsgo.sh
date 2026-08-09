#!/usr/bin/env bash
# Bump the pinned typescript-go version and regenerate the shim layer.
#
# Usage: tools/bump-tsgo.sh [ref]
#   ref: a typescript-go branch, tag, or commit (default: main)
#
# The fixture suite is the compatibility oracle — run `go test ./...` after.
set -euo pipefail
cd "$(dirname "$0")/.."

ref="${1:-main}"

echo "Resolving github.com/microsoft/typescript-go@${ref}..."
version=$(go list -m -json "github.com/microsoft/typescript-go@${ref}" | go run ./tools/internal/jsonfield Version)
echo "Pinning ${version}"

for mod in shim/*/go.mod shim/vfs/*/go.mod; do
    [ -f "$mod" ] || continue
    dir=$(dirname "$mod")
    (cd "$dir" && go mod edit -require="github.com/microsoft/typescript-go@${version}" && go mod tidy >/dev/null)
done

go mod tidy
echo "Regenerating shims..."
go run ./tools/gen_shims
go mod tidy

echo "Done. Now run: go test ./..."
