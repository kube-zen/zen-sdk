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

package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Logger wraps zap.Logger with component-specific context
type Logger struct {
	*zap.Logger
	componentName string
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(err error, msg string, fields ...zap.Field) {
	l.Logger.Error(msg, append(fields, zap.Error(err))...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// NewLogger creates a new structured logger for a component
func NewLogger(componentName string) *Logger {
	// Use controller-runtime's zap logger configuration
	opts := ctrlzap.Options{
		Development: isDevelopment(),
		EncoderConfigOptions: []ctrlzap.EncoderConfigOption{
			func(config *zapcore.EncoderConfig) {
				config.EncodeTime = zapcore.ISO8601TimeEncoder
				config.EncodeLevel = zapcore.CapitalLevelEncoder
			},
		},
	}
	
	zapLogger := ctrlzap.New(ctrlzap.UseFlagOptions(&opts))
	ctrl.SetLogger(zapLogger)
	
	// Get underlying zap.Logger
	underlyingLogger := zapLogger.GetSink().(zap.Sink).(*zap.Logger)
	
	return &Logger{
		Logger:        underlyingLogger,
		componentName: componentName,
	}
}

// WithComponent adds component name to log context
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger:        l.Logger.With(zap.String("component", component)),
		componentName: component,
	}
}

// WithField adds a field to the log context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger:        l.Logger.With(zap.Any(key, value)),
		componentName: l.componentName,
	}
}

// WithFields adds multiple fields to the log context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	
	return &Logger{
		Logger:        l.Logger.With(zapFields...),
		componentName: l.componentName,
	}
}

// isDevelopment checks if we're in development mode
func isDevelopment() bool {
	return os.Getenv("LOG_LEVEL") == "debug" || 
		   os.Getenv("DEVELOPMENT") == "true" ||
		   os.Getenv("ENV") == "development"
}

