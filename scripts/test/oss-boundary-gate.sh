#!/bin/bash
# OSS Boundary Enforcement Gate
# Fails if SaaS-only code patterns are found in OSS repositories

set -e

REPO_ROOT="${1:-$(pwd)}"
FAILED=0
VIOLATIONS=()

# Strict mode: includes _test.go files and optionally scripts/
STRICT_MODE="${OSS_BOUNDARY_STRICT:-0}"

if [ "$STRICT_MODE" = "1" ]; then
	echo "Checking OSS boundary in $REPO_ROOT (STRICT MODE)..."
	echo "Scanning: cmd/, pkg/, internal/ (including _test.go, optionally scripts/)"
	echo "Excluding: docs/, vendor/, dist/"
else
	echo "Checking OSS boundary in $REPO_ROOT..."
	echo "Scanning: cmd/, pkg/, internal/ (excluding scripts/, tests/, fixtures/, docs/, examples/, vendor/, dist/)"
fi
echo ""

# Function to add violation with rule ID
add_violation() {
	local rule_id=$1
	local file=$2
	local line=$3
	local pattern=$4
	local hint=$5
	VIOLATIONS+=("$rule_id|$file|$line|$pattern|$hint")
	FAILED=1
}

# Find relevant Go files (cmd/, pkg/, internal/ only)
SCAN_DIRS=("cmd" "pkg" "internal")
if [ "$STRICT_MODE" = "1" ]; then
	# Strict mode: include _test.go, optionally include scripts/
	GO_FILES=$(find "$REPO_ROOT" -type f -name "*.go" | \
		grep -E "/(cmd|pkg|internal)/" | \
		grep -vE "/(tests?|fixtures?|docs|examples|vendor|dist)/" || true)
	SCRIPT_FILES=$(find "$REPO_ROOT/scripts" -type f -name "*.go" 2>/dev/null || true)
else
	# Default mode: exclude _test.go and scripts/
	GO_FILES=$(find "$REPO_ROOT" -type f -name "*.go" | \
		grep -E "/(cmd|pkg|internal)/" | \
		grep -vE "/(scripts|tests?|fixtures?|docs|examples|vendor|dist)/" | \
		grep -v "_test\.go$" || true)
	SCRIPT_FILES=""
fi

if [ -z "$GO_FILES" ]; then
	echo "⚠️  No Go files found in cmd/, pkg/, internal/"
	exit 0
fi

# OSS001: ZEN_API_BASE_URL references
echo "Checking OSS001: ZEN_API_BASE_URL references..."
while IFS= read -r file; do
	if grep -n "ZEN_API_BASE_URL" "$file" 2>/dev/null | grep -v "oss-boundary-gate.sh" | grep -v "OSS_BOUNDARY.md"; then
		line=$(grep -n "ZEN_API_BASE_URL" "$file" | head -1 | cut -d: -f1)
		add_violation "OSS001" "$file" "$line" "ZEN_API_BASE_URL" "Remove SaaS API base URL; use kubeconfig for OSS operations"
	fi
done <<< "$GO_FILES"

# OSS002: SaaS API endpoint references (/v1/audit, /v1/clusters, /v1/adapters, /v1/tenants)
echo "Checking OSS002: SaaS API endpoint references..."
while IFS= read -r file; do
	if grep -nE "/v1/(audit|clusters|adapters|tenants)" "$file" 2>/dev/null; then
		line=$(grep -nE "/v1/(audit|clusters|adapters|tenants)" "$file" | head -1 | cut -d: -f1)
		pattern=$(grep -nE "/v1/(audit|clusters|adapters|tenants)" "$file" | head -1 | cut -d: -f2- | sed 's/^[[:space:]]*//')
		add_violation "OSS002" "$file" "$line" "$pattern" "Remove SaaS API endpoint; OSS CLI should use Kubernetes APIs only"
	fi
done <<< "$GO_FILES"

# OSS003: src/saas/ imports
echo "Checking OSS003: src/saas/ imports..."
while IFS= read -r file; do
	if grep -n '".*src/saas/' "$file" 2>/dev/null; then
		line=$(grep -n '".*src/saas/' "$file" | head -1 | cut -d: -f1)
		pattern=$(grep -n '".*src/saas/' "$file" | head -1 | cut -d: -f2- | sed 's/^[[:space:]]*//')
		add_violation "OSS003" "$file" "$line" "$pattern" "Remove SaaS package import; use OSS SDK packages only"
	fi
done <<< "$GO_FILES"

# OSS004: Tenant/entitlement SaaS handlers (paired pattern)
echo "Checking OSS004: Tenant/entitlement SaaS handler patterns..."
while IFS= read -r file; do
	if grep -n -iE "tenant.*entitlement|entitlement.*tenant" "$file" 2>/dev/null | grep -v "//" | grep -v "OSS_BOUNDARY.md"; then
		line=$(grep -n -iE "tenant.*entitlement|entitlement.*tenant" "$pattern" "$file" | head -1 | cut -d: -f1)
		pattern=$(grep -n -iE "tenant.*entitlement|entitlement.*tenant" "$file" | head -1 | cut -d: -f2- | sed 's/^[[:space:]]*//')
		add_violation "OSS004" "$file" "$line" "$pattern" "Remove tenant/entitlement SaaS handler; OSS uses K8s CRD status only"
	fi
done <<< "$GO_FILES"

# OSS005: Redis/Cockroach client usage in CLI paths
echo "Checking OSS005: Redis/Cockroach client usage in CLI..."
CLI_GO_FILES=$(echo "$GO_FILES" | grep "/cmd/" || true)
while IFS= read -r file; do
	if grep -n -iE "\bredis\b|\bcockroach\b" "$file" 2>/dev/null | grep -v "test" | grep -v "//"; then
		line=$(grep -n -iE "\bredis\b|\bcockroach\b" "$file" | head -1 | cut -d: -f1)
		pattern=$(grep -n -iE "\bredis\b|\bcockroach\b" "$file" | head -1 | cut -d: -f2- | sed 's/^[[:space:]]*//')
		add_violation "OSS005" "$file" "$line" "$pattern" "Remove Redis/Cockroach client; OSS CLI should not use external databases"
	fi
done <<< "$CLI_GO_FILES"

# OSS006: Committed binaries (ELF, Mach-O, PE executables)
echo "Checking OSS006: Committed binaries..."
# Allowlist: explicitly allowed binaries (none for zen-sdk)
ALLOWLIST_PATTERNS=()
# Find executable files and large files that might be binaries
BINARY_CANDIDATES=$(find "$REPO_ROOT" -type f \( -perm -111 -o -size +1048576 \) 2>/dev/null | \
	grep -vE "/(\.git|vendor|dist|node_modules|\.venv|\.terraform)/" || true)

while IFS= read -r file; do
	# Skip allowlisted files
	skip=0
	for pattern in "${ALLOWLIST_PATTERNS[@]}"; do
		if [[ "$file" == *"$pattern"* ]]; then
			skip=1
			break
		fi
	done
	if [ $skip -eq 1 ]; then
		continue
	fi
	
	# Detect binary file types
	file_type=$(file -b "$file" 2>/dev/null || echo "")
	if echo "$file_type" | grep -qE "ELF|Mach-O|PE|executable|shared object"; then
		# Check for SaaS markers in binary content
		has_saas_marker=0
		if command -v strings >/dev/null 2>&1; then
			binary_content=$(strings -n 8 "$file" 2>/dev/null || echo "")
			if echo "$binary_content" | grep -qE "ZEN_API_BASE_URL|/v1/audit|tenant.*entitlement|entitlement.*tenant"; then
				has_saas_marker=1
				pattern=$(echo "$binary_content" | grep -E "ZEN_API_BASE_URL|/v1/audit|tenant.*entitlement|entitlement.*tenant" | head -1)
			fi
		fi
		
		rel_file="${file#$REPO_ROOT/}"
		if [ $has_saas_marker -eq 1 ]; then
			add_violation "OSS006" "$file" "0" "$pattern" "Binary file contains SaaS markers; zen-sdk must not ship binaries"
		else
			add_violation "OSS006" "$file" "0" "$file_type" "Committed binary file; zen-sdk must not ship binaries (add to .gitignore)"
		fi
	fi
done <<< "$BINARY_CANDIDATES"

# Report violations
if [ $FAILED -eq 0 ]; then
	echo "✅ PASS: OSS boundary check passed"
	exit 0
else
	echo ""
	echo "❌ FAIL: OSS boundary violations detected"
	echo ""
	echo "Violations:"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	for violation in "${VIOLATIONS[@]}"; do
		IFS='|' read -r rule_id file line pattern hint <<< "$violation"
		rel_file="${file#$REPO_ROOT/}"
		echo "Rule: $rule_id"
		echo "  File: $rel_file:$line"
		echo "  Pattern: $pattern"
		echo "  Hint: $hint"
		echo ""
	done
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	exit 1
fi
