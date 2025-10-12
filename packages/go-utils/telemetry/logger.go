package telemetry

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	output string
	logger *log.Logger
}

func NewLogger(output string, header string) *Logger {
	var l *log.Logger

	// Set default header if none provided
	if header == "" {
		header = "waguri"
	} else {
		header = "waguri][" + header
	}

	switch output {
	case "stdout", "":
		l = log.New(os.Stdout, "["+header+"] ", log.LstdFlags)
	case "stderr":
		l = log.New(os.Stderr, "["+header+"] ", log.LstdFlags)
	default:
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		l = log.New(file, "["+header+"] ", log.LstdFlags)
	}

	return &Logger{output: output, logger: l}
}

func (l *Logger) Info(v ...any) {
	message := l.formatMessage(v...)
	l.logger.Println("[INFO]", message)
}

func (l *Logger) Error(v ...any) {
	message := l.formatMessage(v...)
	l.logger.Println("[ERROR]", message)
}

// formatMessage formats multiple arguments with spaces between them, similar to Python's print
func (l *Logger) formatMessage(v ...any) string {
	if len(v) == 0 {
		return ""
	}

	parts := make([]string, len(v))
	for i, arg := range v {
		parts[i] = fmt.Sprintf("%v", arg)
	}
	return strings.Join(parts, " ")
}
