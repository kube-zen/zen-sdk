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

// Package http provides hardened HTTP client with connection pooling, rate limiting, retry logic, and structured logging.
// This package enables consistent HTTP client patterns across OSS components.
package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kube-zen/zen-sdk/pkg/config"
	"github.com/kube-zen/zen-sdk/pkg/logging"
	"github.com/kube-zen/zen-sdk/pkg/retry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

// Prometheus metrics for HTTP client observability
var (
	httpClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_requests_total",
			Help: "Total HTTP client requests by service, method, and status",
		},
		[]string{"service", "method", "status", "retry"},
	)

	httpClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_client_request_duration_seconds",
			Help:    "HTTP client request duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"service", "method", "status"},
	)

	httpClientRetriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_retries_total",
			Help: "Total HTTP client retry attempts by service and reason",
		},
		[]string{"service", "reason"}, // reason: network_error, status_5xx, timeout
	)

	httpClientRateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_rate_limit_hits_total",
			Help: "Total HTTP client rate limit hits by service",
		},
		[]string{"service"},
	)

	httpClientErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_errors_total",
			Help: "Total HTTP client errors by service and error type",
		},
		[]string{"service", "error_type"}, // error_type: timeout, connection_refused, context_cancelled, etc.
	)
)

// ClientConfig holds configuration for HTTP clients
type ClientConfig struct {
	Timeout               time.Duration
	MaxIdleConns          int
	MaxConnsPerHost       int
	IdleConnTimeout       time.Duration
	DisableKeepAlives     bool
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	// TLS configuration
	TLSInsecureSkipVerify bool
	TLSClientConfig       *tls.Config
	// Retry configuration
	RetryConfig retry.HTTPConfig
	// Rate limiting
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
	// Logging
	LoggingEnabled bool
	ServiceName    string
	// Request/Response middleware
	RequestMiddleware  func(*http.Request) error
	ResponseMiddleware func(*http.Response) error
}

// DefaultClientConfig returns a default HTTP client configuration
// Uses environment variables for configuration with sensible defaults
func DefaultClientConfig() *ClientConfig {
	// Get defaults from environment variables
	maxIdleConns := config.RequireEnvIntWithDefault("HTTP_MAX_IDLE_CONNS", 100)
	maxConnsPerHost := config.RequireEnvIntWithDefault("HTTP_MAX_CONNS_PER_HOST", 10)
	idleConnTimeout := config.RequireEnvDurationWithDefault("HTTP_IDLE_CONN_TIMEOUT", 90*time.Second)
	timeout := config.RequireEnvDurationWithDefault("HTTP_TIMEOUT", 30*time.Second)

	return &ClientConfig{
		Timeout:               timeout,
		MaxIdleConns:          maxIdleConns,
		MaxConnsPerHost:       maxConnsPerHost,
		IdleConnTimeout:       idleConnTimeout,
		DisableKeepAlives:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSInsecureSkipVerify: false,
		RetryConfig:           retry.DefaultHTTPConfig(),
		RateLimitEnabled:      false,
		RateLimitRPS:          10.0,
		RateLimitBurst:        10,
		LoggingEnabled:        true,
		ServiceName:           "zen-sdk-http-client",
		RequestMiddleware:     nil,
		ResponseMiddleware:     nil,
	}
}

// Client wraps http.Client with connection pooling, rate limiting, and structured logging
type Client struct {
	client    *http.Client
	config    *ClientConfig
	limiter   *rate.Limiter
	transport *http.Transport
	service   string
	logger    *logging.Logger
}

// NewClient creates a new hardened HTTP client with proper defaults
func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = DefaultClientConfig()
	}

	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	// Configure TLS
	if config.TLSClientConfig != nil {
		transport.TLSClientConfig = config.TLSClientConfig
	} else if config.TLSInsecureSkipVerify {
		// #nosec G402 - TLS InsecureSkipVerify may be required for development/testing
		// or when connecting to self-signed certificates in controlled environments
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	var limiter *rate.Limiter
	if config.RateLimitEnabled {
		limiter = rate.NewLimiter(rate.Limit(config.RateLimitRPS), config.RateLimitBurst)
	}

	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = "zen-sdk-http-client"
	}

	logger := logging.NewLogger("http-client")
	if serviceName != "" && serviceName != "zen-sdk-http-client" {
		logger = logger.WithComponent(serviceName)
	}

	return &Client{
		client:    client,
		config:    config,
		limiter:   limiter,
		transport: transport,
		service:   serviceName,
		logger:    logger,
	}
}

// Do performs an HTTP request with retry logic, rate limiting, and logging
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	startTime := time.Now()

	// Apply request middleware if configured
	if c.config.RequestMiddleware != nil {
		if err := c.config.RequestMiddleware(req); err != nil {
			return nil, fmt.Errorf("request middleware failed: %w", err)
		}
	}

	// Log request start
	if c.config.LoggingEnabled {
		c.logger.WithContext(ctx).Debug("HTTP request started",
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.Operation("http_request"),
		)
	}

	// Apply rate limiting if enabled
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			httpClientRateLimitHitsTotal.WithLabelValues(c.service).Inc()
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}
	}

	// Read body for retry support (if body exists)
	var bodyBytes []byte
	var bodyErr error
	if req.Body != nil {
		bodyBytes, bodyErr = io.ReadAll(req.Body)
		if bodyErr != nil {
			return nil, fmt.Errorf("failed to read request body: %w", bodyErr)
		}
		req.Body.Close()
		// Recreate body reader for first attempt
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Execute request with retry
	var resp *http.Response
	var lastErr error

	retryCfg := c.config.RetryConfig
	for attempt := 0; attempt < retryCfg.MaxAttempts; attempt++ {
		// Check context cancellation
		if ctx != nil {
			select {
			case <-ctx.Done():
				if resp != nil {
					resp.Body.Close()
				}
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			default:
			}
		}

		// Clone request for retry (recreate body if needed)
		reqClone := req.Clone(ctx)
		if bodyBytes != nil {
			reqClone.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			reqClone.ContentLength = int64(len(bodyBytes))
		}

		// Execute request
		var err error
		resp, err = c.client.Do(reqClone)
		if err != nil {
			lastErr = err
			networkRetryReason := "network_error"

			// Record retry metric
			if attempt > 0 {
				httpClientRetriesTotal.WithLabelValues(c.service, networkRetryReason).Inc()
			}

			// Log retry attempt
			if c.config.LoggingEnabled && attempt > 0 {
				c.logger.WithContext(ctx).Warn("HTTP request retry",
					logging.String("method", req.Method),
					logging.String("url", req.URL.String()),
					logging.Int("attempt", attempt+1),
					logging.String("error", err.Error()),
					logging.Operation("http_retry"),
				)
			}

			// Check if error is retryable
			if !retry.ShouldRetry(err, attempt+1, retryCfg.MaxAttempts) {
				// Record error metric
				errorType := "unknown"
				if err == context.DeadlineExceeded || err == context.Canceled {
					errorType = "timeout"
				} else if strings.Contains(err.Error(), "connection refused") {
					errorType = "connection_refused"
				} else if strings.Contains(err.Error(), "context canceled") {
					errorType = "context_cancelled"
				}
				httpClientErrorsTotal.WithLabelValues(c.service, errorType).Inc()
				if c.config.LoggingEnabled {
					c.logger.WithContext(ctx).Error(err, "HTTP request failed (non-retryable)",
						logging.String("method", req.Method),
						logging.String("url", req.URL.String()),
						logging.Int("attempts", attempt+1),
						logging.Operation("http_request"),
					)
				}
				return nil, err
			}

			// Calculate delay and wait
			if attempt+1 < retryCfg.MaxAttempts {
				delay := retry.CalculateDelay(retryCfg.Config, attempt+1)
				if ctx != nil {
					select {
					case <-ctx.Done():
						return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
					case <-time.After(delay):
					}
				} else {
					time.Sleep(delay)
				}
			}
			continue
		}

		// Check if status code is retryable
		if retry.ShouldRetryHTTP(resp.StatusCode, retryCfg) {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			statusRetryReason := "status_5xx"

			// Record retry metric
			if attempt > 0 {
				httpClientRetriesTotal.WithLabelValues(c.service, statusRetryReason).Inc()
			}

			// Log retry attempt
			if c.config.LoggingEnabled && attempt+1 < retryCfg.MaxAttempts {
				c.logger.WithContext(ctx).Warn("HTTP request retry (status code)",
					logging.String("method", req.Method),
					logging.String("url", req.URL.String()),
					logging.Int("status_code", resp.StatusCode),
					logging.Int("attempt", attempt+1),
					logging.Operation("http_retry"),
				)
			}

			// Retry if we haven't exceeded max attempts
			if attempt+1 < retryCfg.MaxAttempts {
				delay := retry.CalculateDelay(retryCfg.Config, attempt+1)
				if ctx != nil {
					select {
					case <-ctx.Done():
						return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
					case <-time.After(delay):
					}
				} else {
					time.Sleep(delay)
				}
				continue
			}
			// Return last response even if retryable (max attempts exceeded)
			resp.Body = io.NopCloser(bytes.NewReader([]byte{})) // Recreate body reader
		} else {
			// Success or non-retryable status code
			break
		}
	}

	// If we exhausted retries without success
	if resp == nil {
		err := fmt.Errorf("request failed after %d attempts: %w", retryCfg.MaxAttempts, lastErr)
		if c.config.LoggingEnabled {
			c.logger.WithContext(ctx).Error(err, "HTTP request failed (max attempts exceeded)",
				logging.String("method", req.Method),
				logging.String("url", req.URL.String()),
				logging.Int("attempts", retryCfg.MaxAttempts),
				logging.Operation("http_request"),
			)
		}
		return nil, err
	}

	// Apply response middleware if configured
	if c.config.ResponseMiddleware != nil {
		if err := c.config.ResponseMiddleware(resp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("response middleware failed: %w", err)
		}
	}

	// Record metrics
	duration := time.Since(startTime)
	statusStr := strconv.Itoa(resp.StatusCode)
	retryStr := "false"
	if lastErr != nil {
		retryStr = "true"
	}

	httpClientRequestsTotal.WithLabelValues(c.service, req.Method, statusStr, retryStr).Inc()
	httpClientRequestDuration.WithLabelValues(c.service, req.Method, statusStr).Observe(duration.Seconds())

	// Log successful request
	if c.config.LoggingEnabled {
		c.logger.WithContext(ctx).Info("HTTP request completed",
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.Int("status_code", resp.StatusCode),
			logging.Duration("duration", duration),
			logging.Operation("http_request"),
		)
	}

	return resp, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post performs a POST request with retry logic
func (c *Client) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// PostJSON performs a POST request with JSON body
func (c *Client) PostJSON(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(req)
}

// Put performs a PUT request with retry logic
func (c *Client) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// Delete performs a DELETE request with retry logic
func (c *Client) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Client returns the underlying http.Client (for compatibility)
func (c *Client) Client() *http.Client {
	return c.client
}

// Close closes idle connections
func (c *Client) Close() {
	c.CloseIdleConnections()
}

// CloseIdleConnections closes all idle connections
func (c *Client) CloseIdleConnections() {
	if c.transport != nil {
		c.transport.CloseIdleConnections()
	}
}

// HardenedHTTPClient is an alias for Client for backward compatibility
// Deprecated: Use Client instead. This alias will be removed in a future version.
type HardenedHTTPClient = Client

// NewHardenedHTTPClient creates a new hardened HTTP client (alias for NewClient)
// Deprecated: Use NewClient instead. This function will be removed in a future version.
func NewHardenedHTTPClient(config *ClientConfig) *HardenedHTTPClient {
	return NewClient(config)
}
