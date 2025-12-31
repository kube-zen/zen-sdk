#!/bin/bash
# Copyright 2025 Kube-ZEN Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# H114: CI guard to prevent re-duplication of shared capabilities
# Fails if shared capabilities are re-implemented in component repos

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "H114: Check for Banned Package Paths (Re-Duplication Guard)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Banned package paths (capabilities that must live in zen-sdk)
# Format: "path:capability:zen-sdk-location"
BANNED_PATHS=(
    "internal/gc:GC logic:zen-sdk/pkg/gc"
    "pkg/gc:GC logic:zen-sdk/pkg/gc"
    "pkg/ratelimiter:Rate limiting:zen-sdk/pkg/gc/ratelimiter"
    "pkg/backoff:Backoff logic:zen-sdk/pkg/gc/backoff"
    "pkg/fieldpath:Field path evaluation:zen-sdk/pkg/gc/fieldpath"
    "pkg/ttl:TTL evaluation:zen-sdk/pkg/gc/ttl"
    "pkg/selector:Selector matching:zen-sdk/pkg/gc/selector"
)

FAILED=0
VIOLATIONS=()

# Check each component repo
for component_dir in "${REPO_ROOT}"/*; do
    if [ ! -d "${component_dir}" ] || [ -L "${component_dir}" ]; then
        continue
    fi

    component_name=$(basename "${component_dir}")
    
    # Skip non-component directories
    if [[ "${component_name}" =~ ^\. ]] || \
       [[ "${component_name}" == "docs" ]] || \
       [[ "${component_name}" == "scripts" ]] || \
       [[ "${component_name}" == "zen-sdk" ]] || \
       [[ "${component_name}" == "zen-admin" ]]; then
        continue
    fi

    # Skip if not a git repo
    if [ ! -d "${component_dir}/.git" ]; then
        continue
    fi

    echo "${BLUE}Checking ${component_name}...${NC}"

    for banned in "${BANNED_PATHS[@]}"; do
        IFS=':' read -r banned_path capability zen_sdk_location <<< "${banned}"
        
        # Search for banned path in component
        if [ -d "${component_dir}/${banned_path}" ]; then
            # Check if it's test code or examples (allowed)
            if find "${component_dir}/${banned_path}" -name "*_test.go" -o -name "example*" | grep -q .; then
                # Has test/example files - check if it also has non-test implementation
                if find "${component_dir}/${banned_path}" -name "*.go" ! -name "*_test.go" ! -name "example*" | grep -q .; then
                    echo "  ${RED}❌ Found banned path: ${banned_path}${NC}"
                    echo "     Capability: ${capability}"
                    echo "     Must use: ${zen_sdk_location}"
                    FAILED=1
                    VIOLATIONS+=("${component_name}:${banned_path}")
                fi
            else
                # No test files - this is likely an implementation
                echo "  ${RED}❌ Found banned path: ${banned_path}${NC}"
                echo "     Capability: ${capability}"
                echo "     Must use: ${zen_sdk_location}"
                FAILED=1
                VIOLATIONS+=("${component_name}:${banned_path}")
            fi
        fi
    done
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ ${FAILED} -eq 0 ]; then
    echo "${GREEN}✅ No banned package paths found${NC}"
    echo "   Shared capabilities stay centralized in zen-sdk"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo "${RED}❌ Found ${#VIOLATIONS[@]} violation(s)${NC}"
    echo ""
    echo "Violations:"
    for violation in "${VIOLATIONS[@]}"; do
        echo "  • ${violation}"
    done
    echo ""
    echo "Shared capabilities must be implemented in zen-sdk, not component repos."
    echo "See zen-sdk/docs/SHARED_CODE_EXTRACTION.md for extraction process."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi

