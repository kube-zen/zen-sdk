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

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient(nil)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.client == nil {
		t.Error("Client.client is nil")
	}
	if client.transport == nil {
		t.Error("Client.transport is nil")
	}
}

func TestNewClient_WithConfig(t *testing.T) {
	config := &ClientConfig{
		Timeout:      10 * time.Second,
		MaxIdleConns: 50,
		ServiceName:  "test-service",
	}
	client := NewClient(config)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.service != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", client.service)
	}
}

func TestClient_Get(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK")) // Ignore error in test handler
	}))
	defer server.Close()

	client := NewClient(nil)
	resp, err := client.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Post(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient(nil)
	resp, err := client.Post(context.Background(), server.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestClient_CloseIdleConnections(t *testing.T) {
	client := NewClient(nil)
	// Should not panic
	client.CloseIdleConnections()
}
