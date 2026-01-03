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

package metadata

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestFilterGitOpsLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected map[string]string
	}{
		{
			name:     "nil labels",
			labels:   nil,
			expected: map[string]string{},
		},
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "no GitOps labels",
			labels: map[string]string{
				"app":     "my-app",
				"version": "1.0.0",
			},
			expected: map[string]string{
				"app":     "my-app",
				"version": "1.0.0",
			},
		},
		{
			name: "with GitOps labels",
			labels: map[string]string{
				"app":                         "my-app",
				"app.kubernetes.io/instance":  "my-instance",
				"argocd.argoproj.io/instance": "my-argocd",
			},
			expected: map[string]string{
				"app": "my-app",
			},
		},
		{
			name: "all GitOps labels",
			labels: map[string]string{
				"app.kubernetes.io/instance":   "my-instance",
				"app.kubernetes.io/managed-by": "helm",
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterGitOpsLabels(tt.labels)
			if len(got) != len(tt.expected) {
				t.Errorf("FilterGitOpsLabels() = %v, want %v", got, tt.expected)
				return
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("FilterGitOpsLabels()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestFilterGitOpsAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    map[string]string
	}{
		{
			name:        "nil annotations",
			annotations: nil,
			expected:    map[string]string{},
		},
		{
			name:        "empty annotations",
			annotations: map[string]string{},
			expected:    map[string]string{},
		},
		{
			name: "no GitOps annotations",
			annotations: map[string]string{
				"description": "my resource",
			},
			expected: map[string]string{
				"description": "my resource",
			},
		},
		{
			name: "with GitOps annotations",
			annotations: map[string]string{
				"description":                  "my resource",
				"argocd.argoproj.io/sync-wave": "1",
			},
			expected: map[string]string{
				"description": "my resource",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterGitOpsAnnotations(tt.annotations)
			if len(got) != len(tt.expected) {
				t.Errorf("FilterGitOpsAnnotations() = %v, want %v", got, tt.expected)
				return
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("FilterGitOpsAnnotations()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestHasOwnerReference(t *testing.T) {
	ownerUID := "owner-uid-123"
	otherUID := "other-uid-456"

	tests := []struct {
		name     string
		obj      metav1.Object
		ownerUID string
		want     bool
	}{
		{
			name: "has owner reference",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{UID: types.UID(ownerUID)},
					},
				},
			},
			ownerUID: ownerUID,
			want:     true,
		},
		{
			name: "no owner reference",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{UID: types.UID(otherUID)},
					},
				},
			},
			ownerUID: ownerUID,
			want:     false,
		},
		{
			name: "empty owner references",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			ownerUID: ownerUID,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasOwnerReference(tt.obj, tt.ownerUID)
			if got != tt.want {
				t.Errorf("HasOwnerReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOwnerReference(t *testing.T) {
	ownerUID := "owner-uid-123"
	otherUID := "other-uid-456"

	tests := []struct {
		name     string
		obj      metav1.Object
		ownerUID string
		want     *metav1.OwnerReference
	}{
		{
			name: "has owner reference",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{UID: types.UID(ownerUID), Name: "owner"},
					},
				},
			},
			ownerUID: ownerUID,
			want:     &metav1.OwnerReference{UID: types.UID(ownerUID), Name: "owner"},
		},
		{
			name: "no matching owner reference",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{UID: types.UID(otherUID)},
					},
				},
			},
			ownerUID: ownerUID,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOwnerReference(tt.obj, tt.ownerUID)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("GetOwnerReference() = %v, want %v", got, tt.want)
				return
			}
			if got != nil && got.UID != tt.want.UID {
				t.Errorf("GetOwnerReference().UID = %q, want %q", got.UID, tt.want.UID)
			}
		})
	}
}

func TestRemoveOwnerReference(t *testing.T) {
	ownerUID := "owner-uid-123"
	otherUID := "other-uid-456"

	tests := []struct {
		name     string
		obj      metav1.Object
		ownerUID string
		want     bool
		wantRefs int
	}{
		{
			name: "remove existing reference",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{UID: types.UID(ownerUID)},
						{UID: types.UID(otherUID)},
					},
				},
			},
			ownerUID: ownerUID,
			want:     true,
			wantRefs: 1,
		},
		{
			name: "remove non-existent reference",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{UID: types.UID(otherUID)},
					},
				},
			},
			ownerUID: ownerUID,
			want:     false,
			wantRefs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveOwnerReference(tt.obj, tt.ownerUID)
			if got != tt.want {
				t.Errorf("RemoveOwnerReference() = %v, want %v", got, tt.want)
			}
			if len(tt.obj.GetOwnerReferences()) != tt.wantRefs {
				t.Errorf("RemoveOwnerReference() left %d refs, want %d", len(tt.obj.GetOwnerReferences()), tt.wantRefs)
			}
		})
	}
}

func TestCopyLabels(t *testing.T) {
	source := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":                        "my-app",
				"app.kubernetes.io/instance": "my-instance",
			},
		},
	}

	target := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{},
	}

	// Test with GitOps filtering
	CopyLabels(source, target, true)
	if target.GetLabels()["app"] != "my-app" {
		t.Errorf("CopyLabels() did not copy 'app' label")
	}
	if target.GetLabels()["app.kubernetes.io/instance"] != "" {
		t.Errorf("CopyLabels() should have filtered GitOps label")
	}

	// Test without GitOps filtering
	target2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{},
	}
	CopyLabels(source, target2, false)
	if target2.GetLabels()["app.kubernetes.io/instance"] != "my-instance" {
		t.Errorf("CopyLabels() should not filter when filterGitOps=false")
	}
}

func TestCopyAnnotations(t *testing.T) {
	source := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"description":                  "my service",
				"argocd.argoproj.io/sync-wave": "1",
			},
		},
	}

	target := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{},
	}

	// Test with GitOps filtering
	CopyAnnotations(source, target, true)
	if target.GetAnnotations()["description"] != "my service" {
		t.Errorf("CopyAnnotations() did not copy 'description' annotation")
	}
	if target.GetAnnotations()["argocd.argoproj.io/sync-wave"] != "" {
		t.Errorf("CopyAnnotations() should have filtered GitOps annotation")
	}
}

func TestSetOwnerReference(t *testing.T) {
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("Failed to add corev1 to scheme: %v", err)
	}

	owner := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "owner",
			UID:  types.UID("owner-uid"),
		},
	}

	dependent := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dependent",
		},
	}

	err := SetOwnerReference(owner, dependent, s)
	if err != nil {
		t.Fatalf("SetOwnerReference() error = %v", err)
	}

	refs := dependent.GetOwnerReferences()
	if len(refs) != 1 {
		t.Fatalf("SetOwnerReference() created %d refs, want 1", len(refs))
	}

	if string(refs[0].UID) != "owner-uid" {
		t.Errorf("SetOwnerReference() UID = %q, want %q", refs[0].UID, "owner-uid")
	}
}

func TestEnsureOwnerReference(t *testing.T) {
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("Failed to add corev1 to scheme: %v", err)
	}

	owner := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "owner",
			UID:  types.UID("owner-uid"),
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
	}

	dependent := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dependent",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
	}

	// First call should add the reference
	added, err := EnsureOwnerReference(owner, dependent, s)
	if err != nil {
		t.Fatalf("EnsureOwnerReference() error = %v", err)
	}
	if !added {
		t.Error("EnsureOwnerReference() should have added reference")
	}

	// Second call should not add (already exists)
	added2, err := EnsureOwnerReference(owner, dependent, s)
	if err != nil {
		t.Fatalf("EnsureOwnerReference() error = %v", err)
	}
	if added2 {
		t.Error("EnsureOwnerReference() should not have added duplicate reference")
	}
}
