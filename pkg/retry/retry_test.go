package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestDo_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3
	// Make generic errors retryable for this test
	config.RetryableErrors = func(err error) bool {
		return err != nil
	}

	attempts := 0
	err := Do(ctx, config, func() error {
		attempts++
		if attempts == 2 {
			return nil
		}
		return errors.New("temporary error")
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3
	// Make errors retryable for this test
	config.RetryableErrors = func(err error) bool {
		return err != nil
	}

	attempts := 0
	err := Do(ctx, config, func() error {
		attempts++
		return errors.New("permanent error")
	})

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultConfig()
	config.MaxAttempts = 10
	config.InitialDelay = 100 * time.Millisecond

	attempts := 0
	done := make(chan error, 1)

	go func() {
		done <- Do(ctx, config, func() error {
			attempts++
			if attempts == 1 {
				cancel()
			}
			return errors.New("error")
		})
	}()

	err := <-done
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestDoWithResult_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3
	// Make errors retryable for this test
	config.RetryableErrors = func(err error) bool {
		return err != nil
	}

	attempts := 0
	result, err := DoWithResult(ctx, config, func() (string, error) {
		attempts++
		if attempts == 2 {
			return "success", nil
		}
		return "", errors.New("temporary error")
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "server timeout",
			err:      k8serrors.NewServerTimeout(schema.GroupResource{Resource: "test"}, "get", 0),
			expected: true,
		},
		{
			name:     "too many requests",
			err:      k8serrors.NewTooManyRequests("test", 0),
			expected: true,
		},
		{
			name:     "conflict",
			err:      k8serrors.NewConflict(schema.GroupResource{Resource: "test"}, "test", errors.New("conflict")),
			expected: true,
		},
		{
			name:     "not found",
			err:      k8serrors.NewNotFound(schema.GroupResource{Resource: "test"}, "test"),
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts=3, got %d", config.MaxAttempts)
	}
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay=100ms, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 5*time.Second {
		t.Errorf("Expected MaxDelay=5s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier=2.0, got %v", config.Multiplier)
	}
	if config.RetryableErrors == nil {
		t.Error("Expected RetryableErrors function, got nil")
	}
}
