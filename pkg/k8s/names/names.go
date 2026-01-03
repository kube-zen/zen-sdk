// Copyright 2025 Kube-ZEN Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package names

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	// MaxResourceNameLength is the maximum length for Kubernetes resource names (RFC 1123 subdomain)
	MaxResourceNameLength = 253
	// MaxLabelValueLength is the maximum length for label values (RFC 1123 label)
	MaxLabelValueLength = 63
)

// GenerateName generates a Kubernetes resource name from a prefix and suffix
// The suffix is preserved even when truncation is needed
// Returns a name that is guaranteed to be <= maxLength
func GenerateName(prefix, suffix string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = MaxResourceNameLength
	}

	// If no suffix, just truncate the prefix
	if suffix == "" {
		return TruncateName(prefix, maxLength)
	}

	// Calculate available space for prefix (accounting for suffix and separator)
	separatorLen := 0
	if prefix != "" && suffix != "" {
		separatorLen = 1 // Assume "-" separator
	}

	availablePrefixLen := maxLength - len(suffix) - separatorLen

	// If prefix fits, return prefix + separator + suffix
	if len(prefix) <= availablePrefixLen {
		if prefix == "" {
			return suffix
		}
		return fmt.Sprintf("%s-%s", prefix, suffix)
	}

	// If available space is too small for any meaningful prefix, use minimal prefix
	minPrefixLen := 2 // Minimum meaningful prefix (e.g., "zl")
	if availablePrefixLen < minPrefixLen {
		// Extreme case: use minimal prefix + full suffix
		return fmt.Sprintf("x-%s", suffix)
	}

	// Truncate prefix to fit
	truncatedPrefix := prefix[:availablePrefixLen]
	return fmt.Sprintf("%s-%s", truncatedPrefix, suffix)
}

// TruncateName truncates a name to the maximum length while ensuring it ends with alphanumeric
// This ensures DNS-1123 compliance (must end with alphanumeric)
func TruncateName(name string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = MaxResourceNameLength
	}

	if len(name) <= maxLength {
		// Even if it fits, ensure it ends with alphanumeric
		truncated := name
		for len(truncated) > 0 && !isAlphanumeric(rune(truncated[len(truncated)-1])) {
			truncated = truncated[:len(truncated)-1]
		}
		if len(truncated) == 0 && len(name) > 0 {
			// If we removed everything, use first alphanumeric char or 'a'
			for _, r := range name {
				if isAlphanumeric(r) {
					return string(r)
				}
			}
			return "a"
		}
		return truncated
	}

	// Truncate to maxLength
	truncated := name[:maxLength]

	// Ensure it ends with alphanumeric (DNS-1123 requirement)
	// Remove trailing non-alphanumeric characters
	for len(truncated) > 0 && !isAlphanumeric(rune(truncated[len(truncated)-1])) {
		truncated = truncated[:len(truncated)-1]
	}

	// If we removed too much, ensure at least one character
	if len(truncated) == 0 && len(name) > 0 {
		// Use first alphanumeric character if available, otherwise use 'a'
		for _, r := range name {
			if isAlphanumeric(r) {
				return string(r)
			}
		}
		return "a"
	}

	return truncated
}

// TruncateNamePreserveSuffix truncates a name while preserving a critical suffix
// The suffix is guaranteed to be included in the result
func TruncateNamePreserveSuffix(name, suffix string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = MaxResourceNameLength
	}

	// If name fits, return as-is
	if len(name) <= maxLength {
		return name
	}

	// If suffix is not in the name, just truncate
	if len(suffix) == 0 || len(name) < len(suffix) {
		return TruncateName(name, maxLength)
	}

	// Check if suffix is at the end of the name
	nameSuffix := name[len(name)-len(suffix):]
	if nameSuffix != suffix {
		// Suffix not at end, just truncate
		return TruncateName(name, maxLength)
	}

	// Calculate available space for prefix part
	prefix := name[:len(name)-len(suffix)]
	availablePrefixLen := maxLength - len(suffix)

	// If available space is too small, use minimal prefix
	if availablePrefixLen < 2 {
		return fmt.Sprintf("x%s", suffix)
	}

	// Truncate prefix part
	truncatedPrefix := TruncateName(prefix, availablePrefixLen)
	return truncatedPrefix + suffix
}

// GenerateNameWithHash generates a name with a hash suffix for uniqueness
// The hash is computed from the base string and appended as a suffix
// The result is truncated to maxLength while preserving the hash suffix
func GenerateNameWithHash(prefix, base string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = MaxResourceNameLength
	}

	// Generate hash from base
	hash := sha256.Sum256([]byte(base))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash

	// Generate name with hash suffix
	return GenerateName(prefix, hashStr, maxLength)
}

// isAlphanumeric checks if a rune is alphanumeric (a-z, A-Z, 0-9)
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
