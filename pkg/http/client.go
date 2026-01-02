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

// Package http provides hardened HTTP client with connection pooling, rate limiting, and structured logging.
// This package enables consistent HTTP client patterns across OSS components.
package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kube-zen/zen-sdk/pkg/config"
	"github.com/kube-zen/zen-sdk/pkg/logging"
	"golang.org/x/time/rate"
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
	// Rate limiting
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
	// Logging
	LoggingEnabled bool
	ServiceName    string
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
		RateLimitEnabled:      false,
		RateLimitRPS:          10.0,
		RateLimitBurst:        10,
		LoggingEnabled:        true,
		ServiceName:           "zen-sdk-http-client",
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

// Do performs an HTTP request with rate limiting and logging
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Apply rate limiting if enabled
	if c.limiter != nil {
		ctx := req.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}
	}

	// Log request if enabled
	if c.config.LoggingEnabled {
		c.logger.Debug("HTTP request",
			logging.Operation("http_request"),
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.String("service", c.service))
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		if c.config.LoggingEnabled {
			c.logger.Warn("HTTP request failed",
				logging.Operation("http_request"),
				logging.Error(err),
				logging.String("method", req.Method),
				logging.String("url", req.URL.String()),
				logging.Duration("duration", duration),
				logging.String("service", c.service))
		}
		return nil, err
	}

	if c.config.LoggingEnabled {
		c.logger.Debug("HTTP response",
			logging.Operation("http_response"),
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.Int("status_code", resp.StatusCode),
			logging.Duration("duration", duration),
			logging.String("service", c.service))
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

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// CloseIdleConnections closes all idle connections
func (c *Client) CloseIdleConnections() {
	if c.transport != nil {
		c.transport.CloseIdleConnections()
	}
}
