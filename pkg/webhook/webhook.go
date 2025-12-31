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

package webhook

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// Patch represents a JSON patch operation
type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// GeneratePatch generates a JSON patch from an object and updates
// Returns a JSON patch as []byte that can be used with admission.Response
func GeneratePatch(obj runtime.Object, updates map[string]interface{}) ([]byte, error) {
	// Convert to unstructured for easier manipulation
	_, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to unstructured: %w", err)
	}

	patches := []Patch{}

	// Generate patches for each update
	for path, value := range updates {
		patches = append(patches, Patch{
			Op:    "replace",
			Path:  path,
			Value: value,
		})
	}

	// Marshal to JSON
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}

	return patchBytes, nil
}

// GenerateAddPatch generates a JSON patch to add a field
func GenerateAddPatch(path string, value interface{}) ([]byte, error) {
	patch := Patch{
		Op:    "add",
		Path:  path,
		Value: value,
	}

	patchBytes, err := json.Marshal([]Patch{patch})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}

	return patchBytes, nil
}

// GenerateRemovePatch generates a JSON patch to remove a field
func GenerateRemovePatch(path string) ([]byte, error) {
	patch := Patch{
		Op:   "remove",
		Path: path,
	}

	patchBytes, err := json.Marshal([]Patch{patch})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}

	return patchBytes, nil
}

// ValidateTLSSecret validates that a secret contains TLS certificate data
func ValidateTLSSecret(secret *unstructured.Unstructured) error {
	data, found, err := unstructured.NestedMap(secret.Object, "data")
	if !found || err != nil {
		return fmt.Errorf("secret data not found or invalid: %w", err)
	}

	requiredKeys := []string{"tls.crt", "tls.key"}
	for _, key := range requiredKeys {
		if _, exists := data[key]; !exists {
			return fmt.Errorf("required TLS key '%s' not found in secret", key)
		}
	}

	return nil
}

// GetNamespacedName extracts NamespacedName from an object
func GetNamespacedName(obj runtime.Object) (types.NamespacedName, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return types.NamespacedName{}, fmt.Errorf("failed to convert object: %w", err)
	}

	name, found, err := unstructured.NestedString(unstructuredObj, "metadata", "name")
	if !found || err != nil {
		return types.NamespacedName{}, fmt.Errorf("object name not found: %w", err)
	}

	namespace, found, err := unstructured.NestedString(unstructuredObj, "metadata", "namespace")
	if !found || err != nil {
		namespace = "" // Namespace is optional
	}

	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, nil
}
