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
	"testing"
)

func TestGenerateName(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		suffix    string
		maxLength int
		want      string
	}{
		{"simple case", "my-prefix", "suffix", 253, "my-prefix-suffix"},
		{"fits within limit", "prefix", "hash123", 253, "prefix-hash123"},
		{"needs truncation", "very-long-prefix-that-exceeds-limit", "hash123", 20, "very-long-prefix-hash123"},
		{"extreme truncation", "very-long-prefix", "hash123", 10, "x-hash123"},
		{"no prefix", "", "suffix", 253, "suffix"},
		{"no suffix", "prefix", "", 253, "prefix"},
		{"both empty", "", "", 253, ""},
		{"default max length", "prefix", "suffix", 0, "prefix-suffix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateName(tt.prefix, tt.suffix, tt.maxLength)
			if len(got) > tt.maxLength && tt.maxLength > 0 {
				t.Errorf("GenerateName(%q, %q, %d) = %q (length %d), exceeds maxLength %d", tt.prefix, tt.suffix, tt.maxLength, got, len(got), tt.maxLength)
			}
			if tt.suffix != "" && len(got) > 0 {
				// Suffix should be preserved
				if len(got) < len(tt.suffix) || got[len(got)-len(tt.suffix):] != tt.suffix {
					t.Errorf("GenerateName(%q, %q, %d) = %q, suffix not preserved", tt.prefix, tt.suffix, tt.maxLength, got)
				}
			}
		})
	}
}

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		want      string
		checkLen  bool
	}{
		{"fits within limit", "my-resource", 253, "my-resource", true},
		{"needs truncation", "very-long-resource-name-that-exceeds-limit", 20, "very-long-resource-n", true},
		{"ends with dash", "my-resource-", 253, "my-resource", true},
		{"ends with dot", "my.resource.", 253, "my.resource", true},
		{"default max length", "my-resource", 0, "my-resource", true},
		{"empty string", "", 253, "", true},
		{"single char", "a", 253, "a", true},
		{"starts with non-alphanumeric", "-resource", 253, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateName(tt.input, tt.maxLength)
			if tt.checkLen && len(got) > tt.maxLength && tt.maxLength > 0 {
				t.Errorf("TruncateName(%q, %d) = %q (length %d), exceeds maxLength %d", tt.input, tt.maxLength, got, len(got), tt.maxLength)
			}
			// Should end with alphanumeric
			if len(got) > 0 {
				lastChar := rune(got[len(got)-1])
				if !isAlphanumeric(lastChar) {
					t.Errorf("TruncateName(%q, %d) = %q, does not end with alphanumeric", tt.input, tt.maxLength, got)
				}
			}
		})
	}
}

func TestTruncateNamePreserveSuffix(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		suffix      string
		maxLength   int
		want        string
		checkSuffix bool
	}{
		{"fits within limit", "prefix-suffix", "suffix", 253, "prefix-suffix", true},
		{"needs truncation", "very-long-prefix-suffix", "suffix", 20, "very-long-prefix-suffix", true},
		{"extreme truncation", "very-long-prefix-suffix", "suffix", 10, "xsuffix", true},
		{"suffix not at end", "prefix-middle-suffix", "suffix", 253, "prefix-middle-suffix", true},
		{"suffix not in name", "prefix-other", "suffix", 253, "prefix-other", false},
		{"empty suffix", "prefix-suffix", "", 20, "prefix-suffix", false},
		{"default max length", "prefix-suffix", "suffix", 0, "prefix-suffix", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateNamePreserveSuffix(tt.input, tt.suffix, tt.maxLength)
			if len(got) > tt.maxLength && tt.maxLength > 0 {
				t.Errorf("TruncateNamePreserveSuffix(%q, %q, %d) = %q (length %d), exceeds maxLength %d", tt.input, tt.suffix, tt.maxLength, got, len(got), tt.maxLength)
			}
			if tt.checkSuffix && tt.suffix != "" && len(got) >= len(tt.suffix) {
				// Suffix should be preserved at the end
				if got[len(got)-len(tt.suffix):] != tt.suffix {
					t.Errorf("TruncateNamePreserveSuffix(%q, %q, %d) = %q, suffix not preserved at end", tt.input, tt.suffix, tt.maxLength, got)
				}
			}
		})
	}
}

func TestGenerateNameWithHash(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		base      string
		maxLength int
		checkLen  bool
	}{
		{"simple case", "zen-lock-inject", "namespace-pod", 253, true},
		{"needs truncation", "very-long-prefix", "base", 20, true},
		{"default max length", "prefix", "base", 0, true},
		{"empty prefix", "", "base", 253, true},
		{"empty base", "prefix", "", 253, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateNameWithHash(tt.prefix, tt.base, tt.maxLength)
			if tt.checkLen && len(got) > tt.maxLength && tt.maxLength > 0 {
				t.Errorf("GenerateNameWithHash(%q, %q, %d) = %q (length %d), exceeds maxLength %d", tt.prefix, tt.base, tt.maxLength, got, len(got), tt.maxLength)
			}
			// Hash should be deterministic
			got2 := GenerateNameWithHash(tt.prefix, tt.base, tt.maxLength)
			if got != got2 {
				t.Errorf("GenerateNameWithHash(%q, %q, %d) = %q, second call = %q, not deterministic", tt.prefix, tt.base, tt.maxLength, got, got2)
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"lowercase letter", 'a', true},
		{"uppercase letter", 'A', true},
		{"digit", '0', true},
		{"dash", '-', false},
		{"dot", '.', false},
		{"underscore", '_', false},
		{"space", ' ', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAlphanumeric(tt.r)
			if got != tt.want {
				t.Errorf("isAlphanumeric(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}
