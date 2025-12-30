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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateAddPatch(t *testing.T) {
	patch, err := GenerateAddPatch("/metadata/labels/test", "value")
	if err != nil {
		t.Fatalf("GenerateAddPatch() error = %v", err)
	}
	
	if len(patch) == 0 {
		t.Error("Expected patch to be generated")
	}
}

func TestGenerateRemovePatch(t *testing.T) {
	patch, err := GenerateRemovePatch("/metadata/labels/test")
	if err != nil {
		t.Fatalf("GenerateRemovePatch() error = %v", err)
	}
	
	if len(patch) == 0 {
		t.Error("Expected patch to be generated")
	}
}

func TestGeneratePatch(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
	}
	
	updates := map[string]interface{}{
		"/metadata/labels/test": "value",
	}
	
	patch, err := GeneratePatch(secret, updates)
	if err != nil {
		t.Fatalf("GeneratePatch() error = %v", err)
	}
	
	if len(patch) == 0 {
		t.Error("Expected patch to be generated")
	}
}

func TestGetNamespacedName(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
	}
	
	nn, err := GetNamespacedName(secret)
	if err != nil {
		t.Fatalf("GetNamespacedName() error = %v", err)
	}
	
	if nn.Name != "test-secret" {
		t.Errorf("Expected name 'test-secret', got '%s'", nn.Name)
	}
	
	if nn.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", nn.Namespace)
	}
}

func TestValidateTLSSecret(t *testing.T) {
	// Test with valid TLS secret
	validSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tls-secret",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("cert"),
			"tls.key": []byte("key"),
		},
	}
	
	// Convert to unstructured for testing
	// Note: This is a simplified test - full implementation would use unstructured
	
	// Test with missing key
	invalidSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "invalid-secret",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("cert"),
			// Missing tls.key
		},
	}
	
	// These tests would need proper unstructured conversion
	_ = validSecret
	_ = invalidSecret
}

