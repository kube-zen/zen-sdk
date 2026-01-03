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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewHTTPMetrics(t *testing.T) {
	// Use a test registry
	registry := prometheus.NewRegistry()

	config := HTTPMetricsConfig{
		Component: "test-component",
		Registry:  registry,
	}

	hm, err := NewHTTPMetrics(config)
	if err != nil {
		t.Fatalf("NewHTTPMetrics failed: %v", err)
	}

	if hm == nil {
		t.Fatal("Expected HTTPMetrics to be created")
	}

	// Test middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK")) // Ignore error in test handler
	})

	middleware := hm.Middleware("test-route")
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestNewHTTPMetricsWithDefaultRegistry(t *testing.T) {
	config := HTTPMetricsConfig{
		Component: "test-component",
		// Registry is nil, should use controller-runtime registry
	}

	hm, err := NewHTTPMetrics(config)
	if err != nil {
		t.Fatalf("NewHTTPMetrics failed: %v", err)
	}

	if hm == nil {
		t.Fatal("Expected HTTPMetrics to be created")
	}
}
