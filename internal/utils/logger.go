// Package utils provides utility functions for the multi-agent system
package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level   LogLevel
	verbose bool
}

// NewLogger creates a new logger with the specified level
func NewLogger(levelStr string, verbose bool) *Logger {
	level := parseLogLevel(levelStr)
	return &Logger{
		level:   level,
		verbose: verbose,
	}
}

// parseLogLevel parses a log level string
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.log(DEBUG, format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.log(INFO, format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.log(WARN, format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.log(ERROR, format, args...)
	}
}

// Verbose logs a verbose message (only if verbose mode is enabled)
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.verbose {
		l.logVerbose(format, args...)
	}
}

// log logs a message with the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)

	if level >= ERROR {
		fmt.Fprintln(os.Stderr, logLine)
	} else {
		fmt.Println(logLine)
	}
}

// logVerbose logs a verbose message with special formatting
func (l *Logger) logVerbose(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", timestamp, message)
}

// Global logger instance
var globalLogger *Logger

// InitLogger initializes the global logger
func InitLogger(level string, verbose bool) {
	globalLogger = NewLogger(level, verbose)
}

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	} else {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an info message using the global logger
func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	} else {
		log.Printf("[INFO] "+format, args...)
	}
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	} else {
		log.Printf("[WARN] "+format, args...)
	}
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	} else {
		log.Printf("[ERROR] "+format, args...)
	}
}

// Verbose logs a verbose message using the global logger
func Verbose(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Verbose(format, args...)
	}
}

// InitLogging initializes global logging with default settings
func InitLogging() {
	globalLogger = NewLogger("INFO", false)
}

// SetLogLevel sets the global log level
func SetLogLevel(level string) {
	if globalLogger != nil {
		globalLogger.level = parseLogLevel(level)
	} else {
		globalLogger = NewLogger(level, false)
	}
}
