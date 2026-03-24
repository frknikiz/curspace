#!/usr/bin/env bash

set -euo pipefail

version="$(go env GOVERSION 2>/dev/null || true)"
case "$version" in
	go1.26*|devel*)
		;;
	*)
		echo "curspace modernize requires Go 1.26+ (found: ${version:-unknown})" >&2
		exit 1
		;;
esac

go fix ./...
gofmt -w cmd internal
