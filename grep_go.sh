#!/bin/sh
set -e

chmod +x "$0"

tmpFile=$(mktemp)

(
cd "$(dirname "$0")" &&
go build -o "$tmpFile" ./
)

if [ ! -f "$tmpFile" ]; then
echo "Failed to build grep_go binary" >&2
exit 1
fi

chmod +x "$tmpFile"

$tmpFile "$@"
exit_code=$?

echo "Exit code: $exit_code"

if [ $exit_code -eq 0 ]; then
echo "Success: Pattern matched"
else
echo "Failure: Pattern did not match"
fi

exit $exit_code

