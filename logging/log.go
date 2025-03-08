package logging

var globalLogger *StandardLogger

func InitLogger(config LoggerConfig) {
	globalLogger = NewLogger(config)
}

// Get the global logger instance
func GetLogger() *StandardLogger {
	return globalLogger
}

// logging/logger.go
// Add this helper function
func L() *StandardLogger {
	return GetLogger()
}
