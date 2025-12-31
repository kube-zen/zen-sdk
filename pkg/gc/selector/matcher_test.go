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

package selector

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestMatchesLabels(t *testing.T) {
	resource := &unstructured.Unstructured{}
	resource.SetLabels(map[string]string{
		"app": "myapp",
		"env": "prod",
	})

	conditions := []LabelCondition{
		{Key: "app", Value: "myapp", Operator: "Equals"},
		{Key: "env", Operator: "Exists"},
	}

	if !MatchesLabels(resource, conditions) {
		t.Error("expected labels to match")
	}
}

func TestMatchesLabels_NotMatch(t *testing.T) {
	resource := &unstructured.Unstructured{}
	resource.SetLabels(map[string]string{
		"app": "myapp",
	})

	conditions := []LabelCondition{
		{Key: "env", Value: "prod", Operator: "Equals"},
	}

	if MatchesLabels(resource, conditions) {
		t.Error("expected labels to not match")
	}
}

func TestMatchesAnnotations(t *testing.T) {
	resource := &unstructured.Unstructured{}
	resource.SetAnnotations(map[string]string{
		"cleanup": "enabled",
	})

	conditions := []AnnotationCondition{
		{Key: "cleanup", Value: "enabled"},
	}

	if !MatchesAnnotations(resource, conditions) {
		t.Error("expected annotations to match")
	}
}

func TestMatchesPhase(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"phase": "Failed",
			},
		},
	}

	phases := []string{"Failed", "Succeeded"}

	if !MatchesPhase(resource, phases) {
		t.Error("expected phase to match")
	}
}

func TestMatchesPhase_NotMatch(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"phase": "Running",
			},
		},
	}

	phases := []string{"Failed", "Succeeded"}

	if MatchesPhase(resource, phases) {
		t.Error("expected phase to not match")
	}
}

func TestMatchesField_Equals(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	condition := FieldCondition{
		Path:     "spec.severity",
		Operator: "Equals",
		Value:    "critical",
	}

	if !MatchesField(resource, condition) {
		t.Error("expected field to match")
	}
}

func TestMatchesField_In(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"severity": "critical",
			},
		},
	}

	condition := FieldCondition{
		Path:     "spec.severity",
		Operator: "In",
		Values:   []string{"critical", "high"},
	}

	if !MatchesField(resource, condition) {
		t.Error("expected field to match")
	}
}

func TestMatchesField_NotIn(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"severity": "low",
			},
		},
	}

	condition := FieldCondition{
		Path:     "spec.severity",
		Operator: "NotIn",
		Values:   []string{"critical", "high"},
	}

	if !MatchesField(resource, condition) {
		t.Error("expected field to match")
	}
}

func TestMatchesConditions(t *testing.T) {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"phase": "Failed",
			},
			"spec": map[string]interface{}{
				"severity": "critical",
			},
		},
	}
	resource.SetLabels(map[string]string{
		"app": "myapp",
	})

	conditions := &Conditions{
		Phase: []string{"Failed", "Succeeded"},
		HasLabels: []LabelCondition{
			{Key: "app", Value: "myapp"},
		},
		And: []FieldCondition{
			{Path: "spec.severity", Operator: "Equals", Value: "critical"},
		},
	}

	if !MatchesConditions(resource, conditions) {
		t.Error("expected all conditions to match")
	}
}

func TestMatchesLabelSelector(t *testing.T) {
	resource := &unstructured.Unstructured{}
	resource.SetLabels(map[string]string{
		"app": "myapp",
		"env": "prod",
	})

	selectorStr := "app=myapp,env=prod"

	matches, err := MatchesLabelSelector(resource, selectorStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matches {
		t.Error("expected label selector to match")
	}
}

func TestMatchesLabelSelector_NotMatch(t *testing.T) {
	resource := &unstructured.Unstructured{}
	resource.SetLabels(map[string]string{
		"app": "myapp",
	})

	selectorStr := "app=myapp,env=prod"

	matches, err := MatchesLabelSelector(resource, selectorStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches {
		t.Error("expected label selector to not match")
	}
}
