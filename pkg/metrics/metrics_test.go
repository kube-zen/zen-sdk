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

package metrics

import (
	"testing"
	"time"
)

func TestNewRecorder(t *testing.T) {
	recorder := NewRecorder("test-component")

	if recorder == nil {
		t.Fatal("Expected recorder to be created")
	}

	if recorder.componentName != "test-component" {
		t.Errorf("Expected component name 'test-component', got '%s'", recorder.componentName)
	}
}

func TestRecordReconciliation(t *testing.T) {
	recorder := NewRecorder("test-component")

	// Record successful reconciliation
	recorder.RecordReconciliation("success", 0.5)

	// Record failed reconciliation
	recorder.RecordReconciliation("error", 1.0)
}

func TestRecordReconciliationSuccess(t *testing.T) {
	recorder := NewRecorder("test-component")

	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	duration := time.Since(start).Seconds()

	recorder.RecordReconciliationSuccess(duration)
}

func TestRecordReconciliationError(t *testing.T) {
	recorder := NewRecorder("test-component")

	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	duration := time.Since(start).Seconds()

	recorder.RecordReconciliationError(duration)
}

func TestRecordError(t *testing.T) {
	recorder := NewRecorder("test-component")

	recorder.RecordError("reconciliation")
	recorder.RecordError("webhook")
}
