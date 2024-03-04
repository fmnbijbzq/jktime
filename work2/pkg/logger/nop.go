package logger

type NopLogger struct {
}

func NewNopLogger() Logger {
	return &NopLogger{}
}

// Debug implements Logger.
func (n *NopLogger) Debug(msg string, args ...Field) {
}

// Error implements Logger.
func (n *NopLogger) Error(msg string, args ...Field) {
}

// Info implements Logger.
func (n *NopLogger) Info(msg string, args ...Field) {
}

// Warn implements Logger.
func (n *NopLogger) Warn(msg string, args ...Field) {
}
