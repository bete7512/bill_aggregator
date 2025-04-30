// pkg/util/logger/logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"encoding/json"
)

// Level represents logging levels
type Level int

const (
	// DEBUG level
	DEBUG Level = iota
	// INFO level
	INFO
	// WARN level
	WARN
	// ERROR level
	ERROR
	// FATAL level
	FATAL
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is a simple logging implementation
type Logger struct {
	Level  Level
	Writer io.Writer
	Format string // "json" or "text"
}

// Config represents the logger configuration
type Config struct {
	Level  string
	Format string
}

// NewLogger creates a new logger
func NewLogger(config Config) *Logger {
	level := parseLevel(config.Level)
	format := strings.ToLower(config.Format)
	if format != "json" && format != "text" {
		format = "json" // Default to JSON format
	}

	return &Logger{
		Level:  level,
		Writer: os.Stdout,
		Format: format,
	}
}

// parseLevel parses the level string to a Level
func parseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO // Default level
	}
}

// WithWriter sets the writer for the logger
func (l *Logger) WithWriter(writer io.Writer) *Logger {
	l.Writer = writer
	return l
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.Level <= DEBUG {
		l.log(DEBUG, msg, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.Level <= INFO {
		l.log(INFO, msg, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.Level <= WARN {
		l.log(WARN, msg, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if l.Level <= ERROR {
		l.log(ERROR, msg, args...)
	}
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(msg string, args ...interface{}) {
	if l.Level <= FATAL {
		l.log(FATAL, msg, args...)
		os.Exit(1)
	}
}

// log performs the actual logging
func (l *Logger) log(level Level, msg string, args ...interface{}) {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	
	// Extract filename from path
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	
	// Get current time
	now := time.Now()
	
	// Convert args to fields
	fields := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				key = fmt.Sprintf("arg%d", i)
			}
			fields[key] = args[i+1]
		}
	}
	
	// Add standard fields
	fields["time"] = now.Format(time.RFC3339)
	fields["level"] = level.String()
	fields["message"] = msg
	fields["caller"] = fmt.Sprintf("%s:%d", file, line)
	
	// Output log based on format
	if l.Format == "json" {
		l.writeJSON(fields)
	} else {
		l.writeText(fields)
	}
}

// writeJSON writes the log in JSON format
func (l *Logger) writeJSON(fields map[string]interface{}) {
	data, err := json.Marshal(fields)
	if err != nil {
		fmt.Fprintf(l.Writer, "{\"level\":\"ERROR\",\"message\":\"Failed to marshal log entry\",\"error\":\"%s\"}\n", err.Error())
		return
	}
	fmt.Fprintln(l.Writer, string(data))
}

// writeText writes the log in text format
func (l *Logger) writeText(fields map[string]interface{}) {
	// Extract standard fields
	time := fields["time"].(string)
	level := fields["level"].(string)
	message := fields["message"].(string)
	caller := fields["caller"].(string)
	
	// Delete standard fields to avoid duplication
	delete(fields, "time")
	delete(fields, "level")
	delete(fields, "message")
	delete(fields, "caller")
	
	// Build the log string
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s [%s] %s (%s)", time, level, message, caller))
	
	// Add remaining fields
	if len(fields) > 0 {
		sb.WriteString(" |")
		for k, v := range fields {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}
	
	fmt.Fprintln(l.Writer, sb.String())
}