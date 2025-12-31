/*
Copyright 2025 Kube-ZEN Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package fieldpath provides field path evaluation primitives for Kubernetes resources.
// Extracted from zen-gc to enable reuse across components.
package fieldpath

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetString retrieves a string value from a resource using a dot-separated field path.
// Example: GetString(resource, "spec.severity") -> ("critical", true, nil)
func GetString(resource *unstructured.Unstructured, path string) (string, bool, error) {
	if path == "" {
		return "", false, fmt.Errorf("field path cannot be empty")
	}

	fieldPath := Parse(path)
	return unstructured.NestedString(resource.Object, fieldPath...)
}

// GetInt64 retrieves an int64 value from a resource using a dot-separated field path.
// Example: GetInt64(resource, "spec.ttlSeconds") -> (3600, true, nil)
func GetInt64(resource *unstructured.Unstructured, path string) (int64, bool, error) {
	if path == "" {
		return 0, false, fmt.Errorf("field path cannot be empty")
	}

	fieldPath := Parse(path)
	return unstructured.NestedInt64(resource.Object, fieldPath...)
}

// GetBool retrieves a boolean value from a resource using a dot-separated field path.
// Example: GetBool(resource, "spec.enabled") -> (true, true, nil)
func GetBool(resource *unstructured.Unstructured, path string) (bool, bool, error) {
	if path == "" {
		return false, false, fmt.Errorf("field path cannot be empty")
	}

	fieldPath := Parse(path)
	return unstructured.NestedBool(resource.Object, fieldPath...)
}

// GetFloat64 retrieves a float64 value from a resource using a dot-separated field path.
// Example: GetFloat64(resource, "spec.threshold") -> (0.95, true, nil)
func GetFloat64(resource *unstructured.Unstructured, path string) (float64, bool, error) {
	if path == "" {
		return 0, false, fmt.Errorf("field path cannot be empty")
	}

	fieldPath := Parse(path)
	return unstructured.NestedFloat64(resource.Object, fieldPath...)
}

// Parse parses a dot-separated field path into a slice for nested field access.
// Example: Parse("spec.severity") -> ["spec", "severity"]
// Example: Parse("status.conditions[0].type") -> ["status", "conditions[0]", "type"]
func Parse(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

// Exists checks if a field path exists in a resource (regardless of value).
// Example: Exists(resource, "spec.severity") -> true
func Exists(resource *unstructured.Unstructured, path string) bool {
	if path == "" {
		return false
	}

	fieldPath := Parse(path)
	_, found, _ := unstructured.NestedFieldNoCopy(resource.Object, fieldPath...)
	return found
}

