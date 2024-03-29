package actions

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Logger is a very basic Github actions compatible logger.
type Logger struct {
	Output io.Writer
}

// NewLogger creates a logger that prints to stdout.
func NewLogger() *Logger {
	return &Logger{Output: os.Stdout}
}

// Debug writes out a debug log message.
func (l *Logger) Debug(message string) {
	fmt.Fprintln(l.Output, "::debug::"+message)
}

// Debugf writes out a formatted debug log message.
func (l *Logger) Debugf(format string, values ...any) {
	fmt.Fprintf(l.Output, "::debug::"+format+"\n", values...)
}

// Info writes out an info log message.
func (l *Logger) Info(message string) {
	fmt.Fprintln(l.Output, message)
}

// Infof writes out a formatted info log message.
func (l *Logger) Infof(format string, values ...any) {
	fmt.Fprintf(l.Output, format+"\n", values...)
}

// Warn writes out a warning log message.
func (l *Logger) Warn(message string) {
	fmt.Fprintln(l.Output, "::warning::"+message)
}

// Warnf writes out a formatted warning log message.
func (l *Logger) Warnf(format string, values ...any) {
	fmt.Fprintf(l.Output, "::warning::"+format+"\n", values...)
}

// Error writes out an error log message.
func (l *Logger) Error(message string) {
	fmt.Fprintln(l.Output, "::error::"+message)
}

// Errorf writes out a formatted error log message.
func (l *Logger) Errorf(format string, values ...any) {
	fmt.Fprintf(l.Output, "::error::"+format+"\n", values...)
}

// SetSecret tells the actions environment to mask the supplied value.
func (l *Logger) SetSecret(value string) {
	fmt.Fprintln(l.Output, "::add-mask::"+value)
}

// GetInputOptions defines some options about how the input should be retrieved.
type GetInputOptions struct {
	Required       bool
	TrimWhitespace bool
}

// GetInput attempts to get the input given the supplied name.
func GetInput(name string, options GetInputOptions) (value string, err error) {
	key := "INPUT_" + strings.ToUpper(regexp.MustCompile(`\s`).ReplaceAllString(name, "_"))
	value = os.Getenv(key)

	if options.Required && value == "" {
		err = fmt.Errorf("input %s is required but was not given", name)
		return
	}

	if options.TrimWhitespace {
		value = strings.TrimSpace(value)
	}

	return
}

// GetMultilineInput gets a multiline input given the supplied name.
func GetMultilineInput(name string, options GetInputOptions) (lines []string, err error) {
	var value string
	value, err = GetInput(name, GetInputOptions{Required: options.Required})
	if err != nil {
		return
	}

	lines = strings.Split(value, "\n")

	if options.TrimWhitespace {
		for i, line := range lines {
			lines[i] = strings.TrimSpace(line)
		}
	}

	return
}
