#!/usr/bin/env bash
# Fail CI if Go sources log Authorization, Signature, idNumber, or dob.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if grep -RIn --include='*.go' --exclude-dir=.git --exclude-dir=vendor \
  -E '(fmt\.Print|log\.(Print|Fatal|Panic)|println)\([^)]*(Authorization|Signature|idNumber|dob|Dob)' .; then
  echo "grep gate failed: sensitive field name appears in a print/log call" >&2
  exit 1
fi

echo "log redaction grep gate passed"
