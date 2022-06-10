package logger

import "go.uber.org/zap"

// Instance used across Lambda invocations for this execution context.
var Instance *zap.Logger

// initialize the logger used across Lambda invocations for the same execution context.
func init() {
	Instance, _ = zap.NewProduction()
}
