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

# H106: CI guard to prevent new imports of zen-sdk/pkg/controller
# Fails if pkg/controller is imported outside the quarantined boundary

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "H106: Check for New pkg/controller Imports"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

FAILED=0

# Allowed locations (quarantined boundary)
ALLOWED_PATHS=(
    "pkg/controller"  # The package itself
    "pkg/controller/REMOVAL_NOTICE.md"  # Removal notice
    "examples"  # Example/migration code
    "docs"  # Documentation
)

# Search for imports of pkg/controller
IMPORTS=$(grep -r "pkg/controller" "${REPO_ROOT}" \
    --include="*.go" \
    --exclude-dir="vendor" \
    --exclude-dir=".git" \
    2>/dev/null | \
    grep -v "REMOVAL_NOTICE\|test\|examples" || true)

if [ -n "${IMPORTS}" ]; then
    echo "${RED}❌ Found imports of pkg/controller outside quarantined boundary:${NC}"
    echo ""
    echo "${IMPORTS}" | while IFS= read -r line; do
        FILE_PATH=$(echo "${line}" | cut -d: -f1)
        CONTENT=$(echo "${line}" | cut -d: -f2-)
        
        # Check if this is an allowed path
        ALLOWED=0
        for allowed_path in "${ALLOWED_PATHS[@]}"; do
            if [[ "${FILE_PATH}" == *"${allowed_path}"* ]]; then
                ALLOWED=1
                break
            fi
        done
        
        if [ ${ALLOWED} -eq 0 ]; then
            echo "  ${YELLOW}${FILE_PATH}${NC}"
            echo "    ${CONTENT}"
            FAILED=1
        fi
    done
    
    if [ ${FAILED} -eq 1 ]; then
        echo ""
        echo "${RED}❌ New imports of pkg/controller detected${NC}"
        echo "   pkg/controller is deprecated and scheduled for removal in v1.0.0"
        echo "   Use zen-sdk/pkg/zenlead instead"
        echo "   See pkg/controller/REMOVAL_NOTICE.md for migration guide"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        exit 1
    fi
fi

echo "${GREEN}✅ No new imports of pkg/controller found${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
exit 0

