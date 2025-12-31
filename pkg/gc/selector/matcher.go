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

// Package selector provides resource selector matching primitives.
// Extracted from zen-gc to enable reuse across components.
package selector

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

// LabelCondition defines a label matching condition.
type LabelCondition struct {
	Key      string
	Value    string
	Operator string // Exists, Equals, In, NotIn
}

// AnnotationCondition defines an annotation matching condition.
type AnnotationCondition struct {
	Key   string
	Value string
}

// FieldCondition defines a field matching condition.
type FieldCondition struct {
	Path     string   // Dot-separated field path (e.g., "status.phase")
	Operator string   // Equals, In, NotIn
	Value    string   // Single value
	Values   []string // Multiple values (for In/NotIn)
}

// Conditions defines a set of matching conditions.
type Conditions struct {
	Phase          []string              // Match status.phase (OR logic)
	HasLabels      []LabelCondition      // Match labels (AND logic)
	HasAnnotations []AnnotationCondition // Match annotations (AND logic)
	And            []FieldCondition      // Match arbitrary fields (AND logic)
}

// MatchesLabels checks if a resource matches label conditions.
// All conditions must match (AND logic).
func MatchesLabels(resource *unstructured.Unstructured, conditions []LabelCondition) bool {
	resourceLabels := resource.GetLabels()
	for _, cond := range conditions {
		value, exists := resourceLabels[cond.Key]
		switch cond.Operator {
		case "Exists", "":
			if !exists {
				return false
			}
		case "Equals":
			if !exists || value != cond.Value {
				return false
			}
		case "In":
			if !exists || value != cond.Value {
				return false
			}
		case "NotIn":
			if exists && value == cond.Value {
				return false
			}
		default:
			// Unknown operator - fail safe
			return false
		}
	}
	return true
}

// MatchesAnnotations checks if a resource matches annotation conditions.
// All conditions must match (AND logic).
func MatchesAnnotations(resource *unstructured.Unstructured, conditions []AnnotationCondition) bool {
	resourceAnnotations := resource.GetAnnotations()
	for _, cond := range conditions {
		value, exists := resourceAnnotations[cond.Key]
		if !exists || value != cond.Value {
			return false
		}
	}
	return true
}

// MatchesPhase checks if a resource's status.phase matches any of the required phases.
// If no phases specified, returns true (OR logic).
func MatchesPhase(resource *unstructured.Unstructured, phases []string) bool {
	if len(phases) == 0 {
		return true
	}
	phase, found, _ := unstructured.NestedString(resource.Object, "status", "phase")
	if !found {
		return false
	}
	for _, p := range phases {
		if phase == p {
			return true
		}
	}
	return false
}

// MatchesField checks if a resource field matches a field condition.
func MatchesField(resource *unstructured.Unstructured, condition FieldCondition) bool {
	// Parse field path
	fieldPath := parseFieldPath(condition.Path)
	fieldValue, found, _ := unstructured.NestedString(resource.Object, fieldPath...)
	if !found {
		return false
	}

	switch condition.Operator {
	case "Equals", "":
		return fieldValue == condition.Value
	case "In":
		// Check if value is in Values list
		for _, v := range condition.Values {
			if fieldValue == v {
				return true
			}
		}
		// Fallback to single Value if Values is empty
		return fieldValue == condition.Value
	case "NotIn":
		// Check if value is NOT in Values list
		for _, v := range condition.Values {
			if fieldValue == v {
				return false
			}
		}
		// Fallback to single Value if Values is empty
		return fieldValue != condition.Value
	default:
		return false
	}
}

// MatchesFields checks if a resource matches all field conditions (AND logic).
func MatchesFields(resource *unstructured.Unstructured, conditions []FieldCondition) bool {
	for _, cond := range conditions {
		if !MatchesField(resource, cond) {
			return false
		}
	}
	return true
}

// MatchesConditions checks if a resource matches all conditions (AND logic).
func MatchesConditions(resource *unstructured.Unstructured, conditions *Conditions) bool {
	if conditions == nil {
		return true
	}

	if !MatchesPhase(resource, conditions.Phase) {
		return false
	}
	if !MatchesLabels(resource, conditions.HasLabels) {
		return false
	}
	if !MatchesAnnotations(resource, conditions.HasAnnotations) {
		return false
	}
	if !MatchesFields(resource, conditions.And) {
		return false
	}
	return true
}

// MatchesLabelSelector checks if a resource matches a Kubernetes label selector.
func MatchesLabelSelector(resource *unstructured.Unstructured, selectorStr string) (bool, error) {
	if selectorStr == "" {
		return true, nil
	}

	selector, err := labels.Parse(selectorStr)
	if err != nil {
		return false, err
	}

	resourceLabels := labels.Set(resource.GetLabels())
	return selector.Matches(resourceLabels), nil
}

// parseFieldPath parses a dot-separated field path into a slice.
func parseFieldPath(path string) []string {
	if path == "" {
		return nil
	}
	// Simple split - can be enhanced for array indexing if needed
	result := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

