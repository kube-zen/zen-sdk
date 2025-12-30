/*
Example: Using zen-sdk/pkg/webhook in an admission webhook

This example shows how to use zen-sdk/pkg/webhook for
webhook operations like patch generation.
*/

package main

import (
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kube-zen/zen-sdk/pkg/webhook"
)

// ExampleMutatingWebhook shows how to use webhook helpers
func ExampleMutatingWebhook(req admission.Request) admission.Response {
	// Decode the object
	secret := &corev1.Secret{}
	if err := req.Decode(secret); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	
	// Generate patch to add a label
	updates := map[string]interface{}{
		"/metadata/labels/managed-by": "zen-sdk",
	}
	
	patch, err := webhook.GeneratePatch(secret, updates)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	
	return admission.Patched("", patch)
}

// ExampleAddLabel shows how to add a label
func ExampleAddLabel() (admission.Patch, error) {
	return webhook.GenerateAddPatch("/metadata/labels/test", "value")
}

// ExampleRemoveLabel shows how to remove a label
func ExampleRemoveLabel() (admission.Patch, error) {
	return webhook.GenerateRemovePatch("/metadata/labels/test")
}

// ExampleGetNamespacedName shows how to extract NamespacedName
func ExampleGetNamespacedName(obj interface{}) error {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return nil
	}
	
	nn, err := webhook.GetNamespacedName(secret)
	if err != nil {
		return err
	}
	
	// Use nn.Name and nn.Namespace
	_ = nn
	return nil
}

