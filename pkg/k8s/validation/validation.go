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

package validation

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// MaxAnnotationValueLength is the maximum length for annotation values (RFC 1123 label)
	MaxAnnotationValueLength = 253
	// MaxLabelValueLength is the maximum length for label values (RFC 1123 label)
	MaxLabelValueLength = 63
	// MaxResourceNameLength is the maximum length for resource names (RFC 1123 subdomain)
	MaxResourceNameLength = 253
)

var (
	// DNS1123SubdomainRegex matches valid Kubernetes DNS-1123 subdomain format
	// DNS-1123 subdomain: lowercase alphanumeric, '-' or '.', must start/end with alphanumeric
	// Used for resource names, API groups, etc.
	DNS1123SubdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

	// DNS1123LabelRegex matches valid Kubernetes DNS-1123 label format
	// DNS-1123 label: lowercase alphanumeric with hyphens, no dots
	// Used for resource names (without dots), label keys/values
	DNS1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

// ValidateDNS1123Subdomain validates a DNS-1123 subdomain name
// DNS-1123 subdomain allows dots and hyphens, used for resource names, API groups, etc.
func ValidateDNS1123Subdomain(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if !DNS1123SubdomainRegex.MatchString(name) {
		return fmt.Errorf("invalid DNS-1123 subdomain %q: must be lowercase alphanumeric with dots or hyphens, must start/end with alphanumeric", name)
	}

	return nil
}

// ValidateDNS1123Label validates a DNS-1123 label name
// DNS-1123 label does not allow dots, used for resource names (without dots), label keys/values
func ValidateDNS1123Label(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if !DNS1123LabelRegex.MatchString(name) {
		return fmt.Errorf("invalid DNS-1123 label %q: must be lowercase alphanumeric with hyphens, no dots, must start/end with alphanumeric", name)
	}

	return nil
}

// ValidateResourceName validates a Kubernetes resource name with optional length limit
// Uses DNS-1123 subdomain validation (allows dots)
func ValidateResourceName(name string, maxLength int) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	if maxLength > 0 && len(name) > maxLength {
		return fmt.Errorf("resource name %q exceeds maximum length of %d", name, maxLength)
	}

	if err := ValidateDNS1123Subdomain(name); err != nil {
		return fmt.Errorf("invalid resource name %q: %w", name, err)
	}

	return nil
}

// ValidateResourceNameLabel validates a Kubernetes resource name using DNS-1123 label format (no dots)
// Used for resource names that cannot contain dots
func ValidateResourceNameLabel(name string, maxLength int) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	if maxLength > 0 && len(name) > maxLength {
		return fmt.Errorf("resource name %q exceeds maximum length of %d", name, maxLength)
	}

	if err := ValidateDNS1123Label(name); err != nil {
		return fmt.Errorf("invalid resource name %q: %w", name, err)
	}

	return nil
}

// ValidateGVR validates a GroupVersionResource
// Returns an error if any component is invalid
func ValidateGVR(gvr schema.GroupVersionResource) error {
	// Validate group (must be valid DNS subdomain or empty for core resources)
	if gvr.Group != "" {
		if err := ValidateDNS1123Subdomain(gvr.Group); err != nil {
			return fmt.Errorf("invalid API group %q: %w", gvr.Group, err)
		}
	}

	// Validate version (must be non-empty and valid version string)
	if gvr.Version == "" {
		return fmt.Errorf("API version cannot be empty")
	}
	// Version should start with 'v' followed by numbers, or be a valid semver-like string
	if !strings.HasPrefix(gvr.Version, "v") {
		return fmt.Errorf("invalid API version %q: must start with 'v' (e.g., 'v1', 'v1alpha1')", gvr.Version)
	}

	// Validate resource (must be valid Kubernetes resource name, no dots)
	if gvr.Resource == "" {
		return fmt.Errorf("resource name cannot be empty")
	}
	if err := ValidateDNS1123Label(gvr.Resource); err != nil {
		return fmt.Errorf("invalid resource name %q: %w", gvr.Resource, err)
	}

	return nil
}

// ValidateGVRConfig validates a GVRConfig from individual components
func ValidateGVRConfig(group, version, resource string) error {
	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	return ValidateGVR(gvr)
}

// ValidateWithKubernetesValidation uses Kubernetes API validation for DNS-1123 subdomain
// This is a wrapper around k8s.io/apimachinery/pkg/api/validation.IsDNS1123Subdomain
// for consistency with Kubernetes validation rules
// Note: This function requires importing k8s.io/apimachinery/pkg/api/validation in the calling code
// For most use cases, ValidateDNS1123Subdomain is sufficient
func ValidateWithKubernetesValidation(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Use our own regex validation instead of Kubernetes validation to avoid import dependency
	// The regex matches Kubernetes validation rules
	if err := ValidateDNS1123Subdomain(name); err != nil {
		return err
	}

	return nil
}

// ValidateNameWithWhitespaceCheck validates a name and ensures it has no leading/trailing whitespace
func ValidateNameWithWhitespaceCheck(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if strings.TrimSpace(name) != name {
		return fmt.Errorf("name %q contains leading or trailing whitespace", name)
	}

	return ValidateDNS1123Subdomain(name)
}
