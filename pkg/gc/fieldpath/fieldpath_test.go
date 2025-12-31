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

package fieldpath

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetString(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	value, found, err := GetString(resource, "spec.severity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected field to be found")
	}
	if value != "critical" {
		t.Errorf("expected 'critical', got '%s'", value)
	}
}

func TestGetString_NotFound(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{},
		},
	}

	_, found, err := GetString(resource, "spec.severity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected field to not be found")
	}
}

func TestGetInt64(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"ttlSeconds": int64(3600),
			},
		},
	}

	value, found, err := GetInt64(resource, "spec.ttlSeconds")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected field to be found")
	}
	if value != 3600 {
		t.Errorf("expected 3600, got %d", value)
	}
}

func TestGetBool(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	value, found, err := GetBool(resource, "spec.enabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected field to be found")
	}
	if !value {
		t.Error("expected true, got false")
	}
}

func TestGetFloat64(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"threshold": 0.95,
			},
		},
	}

	value, found, err := GetFloat64(resource, "spec.threshold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected field to be found")
	}
	if value != 0.95 {
		t.Errorf("expected 0.95, got %f", value)
	}
}

func TestExists(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	if !Exists(resource, "spec.severity") {
		t.Error("expected field to exist")
	}

	if Exists(resource, "spec.nonexistent") {
		t.Error("expected field to not exist")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"spec.severity", []string{"spec", "severity"}},
		{"status.conditions", []string{"status", "conditions"}},
		{"metadata.labels.app", []string{"metadata", "labels", "app"}},
		{"", nil},
	}

	for _, tt := range tests {
		result := Parse(tt.path)
		if len(result) != len(tt.expected) {
			t.Errorf("path %s: expected %v, got %v", tt.path, tt.expected, result)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("path %s: expected %v, got %v", tt.path, tt.expected, result)
				break
			}
		}
	}
}

func TestGetString_EmptyPath(t *testing.T) {
	resource := &unstructured.Unstructured{}

	_, _, err := GetString(resource, "")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

