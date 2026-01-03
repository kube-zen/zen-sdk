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
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestValidateDNS1123Subdomain(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "my-resource", false},
		{"valid with dots", "my.resource.name", false},
		{"valid with numbers", "resource123", false},
		{"valid single char", "a", false},
		{"empty name", "", true},
		{"starts with dash", "-resource", true},
		{"ends with dash", "resource-", true},
		{"starts with dot", ".resource", true},
		{"ends with dot", "resource.", true},
		{"uppercase", "Resource", true},
		{"with underscore", "resource_name", true},
		{"with spaces", "resource name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDNS1123Subdomain(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDNS1123Subdomain(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDNS1123Label(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "my-resource", false},
		{"valid with numbers", "resource123", false},
		{"valid single char", "a", false},
		{"empty name", "", true},
		{"starts with dash", "-resource", true},
		{"ends with dash", "resource-", true},
		{"with dots", "my.resource", true},
		{"uppercase", "Resource", true},
		{"with underscore", "resource_name", true},
		{"with spaces", "resource name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDNS1123Label(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDNS1123Label(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantErr   bool
	}{
		{"valid name", "my-resource", 0, false},
		{"valid with dots", "my.resource.name", 0, false},
		{"valid within length", "my-resource", 253, false},
		{"exceeds length", "my-resource", 10, true},
		{"empty name", "", 0, true},
		{"invalid format", "My-Resource", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceName(tt.input, tt.maxLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceName(%q, %d) error = %v, wantErr %v", tt.input, tt.maxLength, err, tt.wantErr)
			}
		})
	}
}

func TestValidateResourceNameLabel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantErr   bool
	}{
		{"valid name", "my-resource", 0, false},
		{"valid within length", "my-resource", 63, false},
		{"exceeds length", "my-resource", 10, true},
		{"with dots", "my.resource", 0, true},
		{"empty name", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceNameLabel(tt.input, tt.maxLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceNameLabel(%q, %d) error = %v, wantErr %v", tt.input, tt.maxLength, err, tt.wantErr)
			}
		})
	}
}

func TestValidateGVR(t *testing.T) {
	tests := []struct {
		name    string
		gvr     schema.GroupVersionResource
		wantErr bool
	}{
		{
			name: "valid core resource",
			gvr: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			wantErr: false,
		},
		{
			name: "valid custom resource",
			gvr: schema.GroupVersionResource{
				Group:    "apps.kube-zen.io",
				Version:  "v1alpha1",
				Resource: "jobflows",
			},
			wantErr: false,
		},
		{
			name: "empty version",
			gvr: schema.GroupVersionResource{
				Group:    "apps.kube-zen.io",
				Version:  "",
				Resource: "jobflows",
			},
			wantErr: true,
		},
		{
			name: "invalid version format",
			gvr: schema.GroupVersionResource{
				Group:    "apps.kube-zen.io",
				Version:  "1alpha1",
				Resource: "jobflows",
			},
			wantErr: true,
		},
		{
			name: "empty resource",
			gvr: schema.GroupVersionResource{
				Group:    "apps.kube-zen.io",
				Version:  "v1alpha1",
				Resource: "",
			},
			wantErr: true,
		},
		{
			name: "invalid group format",
			gvr: schema.GroupVersionResource{
				Group:    "Apps.Kube-Zen.IO",
				Version:  "v1alpha1",
				Resource: "jobflows",
			},
			wantErr: true,
		},
		{
			name: "invalid resource format",
			gvr: schema.GroupVersionResource{
				Group:    "apps.kube-zen.io",
				Version:  "v1alpha1",
				Resource: "JobFlows",
			},
			wantErr: true,
		},
		{
			name: "resource with dots",
			gvr: schema.GroupVersionResource{
				Group:    "apps.kube-zen.io",
				Version:  "v1alpha1",
				Resource: "job.flows",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGVR(tt.gvr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGVR(%+v) error = %v, wantErr %v", tt.gvr, err, tt.wantErr)
			}
		})
	}
}

func TestValidateGVRConfig(t *testing.T) {
	tests := []struct {
		name     string
		group    string
		version  string
		resource string
		wantErr  bool
	}{
		{"valid", "apps.kube-zen.io", "v1alpha1", "jobflows", false},
		{"empty version", "apps.kube-zen.io", "", "jobflows", true},
		{"invalid version", "apps.kube-zen.io", "1alpha1", "jobflows", true},
		{"empty resource", "apps.kube-zen.io", "v1alpha1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGVRConfig(tt.group, tt.version, tt.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGVRConfig(%q, %q, %q) error = %v, wantErr %v", tt.group, tt.version, tt.resource, err, tt.wantErr)
			}
		})
	}
}

func TestValidateWithKubernetesValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-resource", false},
		{"valid with dots", "my.resource.name", false},
		{"empty name", "", true},
		{"invalid format", "My-Resource", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithKubernetesValidation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithKubernetesValidation(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNameWithWhitespaceCheck(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-resource", false},
		{"leading whitespace", " my-resource", true},
		{"trailing whitespace", "my-resource ", true},
		{"both whitespace", " my-resource ", true},
		{"empty name", "", true},
		{"invalid format", "My-Resource", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNameWithWhitespaceCheck(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNameWithWhitespaceCheck(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
