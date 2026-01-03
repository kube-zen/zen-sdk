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
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GitOpsTrackingLabels contains common GitOps tool labels that should NOT be copied to generated resources
// These labels would cause ownership/prune conflicts when resources are managed by controllers
var GitOpsTrackingLabels = map[string]struct{}{
	"app.kubernetes.io/instance":            {},
	"app.kubernetes.io/managed-by":          {}, // Controllers set their own value
	"app.kubernetes.io/part-of":             {},
	"app.kubernetes.io/version":             {},
	"argocd.argoproj.io/instance":           {},
	"fluxcd.io/part-of":                     {},
	"kustomize.toolkit.fluxcd.io/name":      {},
	"kustomize.toolkit.fluxcd.io/namespace": {},
	"kustomize.toolkit.fluxcd.io/revision":  {},
}

// GitOpsTrackingAnnotations contains common GitOps tool annotations that should NOT be copied to generated resources
// These annotations would cause ownership/prune conflicts when resources are managed by controllers
var GitOpsTrackingAnnotations = map[string]struct{}{
	"argocd.argoproj.io/sync-wave":         {},
	"argocd.argoproj.io/sync-options":      {},
	"fluxcd.io/sync-checksum":              {},
	"kustomize.toolkit.fluxcd.io/checksum": {},
}

// FilterGitOpsLabels removes GitOps tracking labels from a label map
// Optimized: O(n) with map lookup instead of O(n*m) with nested loops
// Returns a new map with GitOps tracking labels removed
func FilterGitOpsLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return make(map[string]string)
	}
	// Pre-allocate with estimated capacity (most labels will pass through)
	filtered := make(map[string]string, len(labels))
	for k, v := range labels {
		if _, skip := GitOpsTrackingLabels[k]; !skip {
			filtered[k] = v
		}
	}
	return filtered
}

// FilterGitOpsAnnotations removes GitOps tracking annotations from an annotation map
// Optimized: O(n) with map lookup instead of O(n*m) with nested loops
// Returns a new map with GitOps tracking annotations removed
func FilterGitOpsAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return make(map[string]string)
	}
	// Pre-allocate with estimated capacity (most annotations will pass through)
	filtered := make(map[string]string, len(annotations))
	for k, v := range annotations {
		if _, skip := GitOpsTrackingAnnotations[k]; !skip {
			filtered[k] = v
		}
	}
	return filtered
}

// SetOwnerReference sets the owner reference on a dependent object
// This is a convenience wrapper around controllerutil.SetControllerReference
// that handles common error cases
func SetOwnerReference(owner, dependent metav1.Object, scheme *runtime.Scheme) error {
	return controllerutil.SetControllerReference(owner, dependent, scheme)
}

// HasOwnerReference checks if an object has an owner reference to the specified owner
func HasOwnerReference(obj metav1.Object, ownerUID string) bool {
	refs := obj.GetOwnerReferences()
	for _, ref := range refs {
		if string(ref.UID) == ownerUID {
			return true
		}
	}
	return false
}

// GetOwnerReference returns the owner reference for the specified owner UID, if it exists
func GetOwnerReference(obj metav1.Object, ownerUID string) *metav1.OwnerReference {
	refs := obj.GetOwnerReferences()
	for _, ref := range refs {
		if string(ref.UID) == ownerUID {
			return &ref
		}
	}
	return nil
}

// RemoveOwnerReference removes the owner reference for the specified owner UID
// Returns true if a reference was removed, false otherwise
func RemoveOwnerReference(obj metav1.Object, ownerUID string) bool {
	refs := obj.GetOwnerReferences()
	for i, ref := range refs {
		if string(ref.UID) == ownerUID {
			// Remove the reference
			refs = append(refs[:i], refs[i+1:]...)
			obj.SetOwnerReferences(refs)
			return true
		}
	}
	return false
}

// EnsureOwnerReference ensures an owner reference exists for the specified owner
// If it doesn't exist, it creates one. If it exists but is different, it updates it.
// Returns true if the owner reference was added or updated, false otherwise
// Note: This function requires owner to be a runtime.Object to get GVK information
func EnsureOwnerReference(owner runtime.Object, dependent metav1.Object, scheme *runtime.Scheme) (bool, error) {
	ownerObj, err := GetAccessor(owner)
	if err != nil {
		return false, err
	}

	existingRef := GetOwnerReference(dependent, string(ownerObj.GetUID()))
	if existingRef != nil {
		// Get GVK from owner
		gvk, err := getGVK(owner, scheme)
		if err != nil {
			return false, err
		}

		// Check if it needs updating
		if existingRef.APIVersion == gvk.GroupVersion().String() &&
			existingRef.Kind == gvk.Kind &&
			existingRef.Name == ownerObj.GetName() &&
			existingRef.Controller != nil && *existingRef.Controller {
			// Already correct, no update needed
			return false, nil
		}
		// Remove old reference
		RemoveOwnerReference(dependent, string(ownerObj.GetUID()))
	}

	// Add new reference
	err = SetOwnerReference(ownerObj, dependent, scheme)
	if err != nil {
		return false, err
	}
	return true, nil
}

// getGVK gets the GroupVersionKind for an object using the scheme
func getGVK(obj runtime.Object, s *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvks, _, err := s.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if len(gvks) == 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("no GVK found for object")
	}
	return gvks[0], nil
}

// CopyLabels copies labels from source to target, optionally filtering GitOps labels
func CopyLabels(source, target metav1.Object, filterGitOps bool) {
	if source == nil || target == nil {
		return
	}

	sourceLabels := source.GetLabels()
	if sourceLabels == nil {
		return
	}

	if filterGitOps {
		sourceLabels = FilterGitOpsLabels(sourceLabels)
	}

	if target.GetLabels() == nil {
		target.SetLabels(make(map[string]string))
	}

	for k, v := range sourceLabels {
		target.GetLabels()[k] = v
	}
}

// CopyAnnotations copies annotations from source to target, optionally filtering GitOps annotations
func CopyAnnotations(source, target metav1.Object, filterGitOps bool) {
	if source == nil || target == nil {
		return
	}

	sourceAnnotations := source.GetAnnotations()
	if sourceAnnotations == nil {
		return
	}

	if filterGitOps {
		sourceAnnotations = FilterGitOpsAnnotations(sourceAnnotations)
	}

	if target.GetAnnotations() == nil {
		target.SetAnnotations(make(map[string]string))
	}

	for k, v := range sourceAnnotations {
		target.GetAnnotations()[k] = v
	}
}

// GetAccessor returns a metav1.Object accessor for the given object
// This is useful when working with unstructured objects
func GetAccessor(obj runtime.Object) (metav1.Object, error) {
	return meta.Accessor(obj)
}
