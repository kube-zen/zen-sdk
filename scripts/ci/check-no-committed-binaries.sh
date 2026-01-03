#!/usr/bin/env bash
# Check for committed binaries in the repository
# Fails if ELF/Mach-O/PE files are found (allowlist only if required)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FAILED=0
BINARIES=()

echo "Checking for committed binaries..."

# Allowlist: explicitly allowed binaries (default: none for zen-sdk)
# Format: relative paths from repo root
ALLOWLIST=(
	# Example: "scripts/tools/helper"
)

# Find files that might be binaries
# Check executable bit OR files > 1MB
CANDIDATES=$(find "$SCRIPT_DIR" -type f \( -perm -111 -o -size +1048576 \) 2>/dev/null | \
	grep -vE "/(\.git|vendor|dist|node_modules|\.venv|\.terraform|testdata)/" || true)

if [ -z "$CANDIDATES" ]; then
	echo "  ✅ No binary candidates found"
	exit 0
fi

while IFS= read -r file; do
	# Skip allowlisted files
	rel_file="${file#$SCRIPT_DIR/}"
	skip=0
	for allowed in "${ALLOWLIST[@]}"; do
		if [[ "$rel_file" == "$allowed" ]] || [[ "$rel_file" == "$allowed"* ]]; then
			skip=1
			break
		fi
	done
	if [ $skip -eq 1 ]; then
		continue
	fi
	
	# Detect binary file types
	if command -v file >/dev/null 2>&1; then
		file_type=$(file -b "$file" 2>/dev/null || echo "")
		if echo "$file_type" | grep -qE "^(ELF|Mach-O|PE|executable|shared object)"; then
			BINARIES+=("$rel_file|$file_type")
			FAILED=1
		fi
	fi
done <<< "$CANDIDATES"

if [ $FAILED -eq 0 ]; then
	echo "  ✅ No committed binaries detected"
	exit 0
else
	echo ""
	echo "  ❌ Committed binaries detected (zen-sdk must not ship binaries):"
	echo ""
	for binary_info in "${BINARIES[@]}"; do
		IFS='|' read -r file file_type <<< "$binary_info"
		echo "    File: $file"
		echo "    Type: $file_type"
		echo "    Remediation: Remove file and add to .gitignore"
		echo ""
	done
	echo "  Hint: zen-sdk is a library-only repo; binaries belong in zen-watcher/cmd/zenctl"
	exit 1
fi

