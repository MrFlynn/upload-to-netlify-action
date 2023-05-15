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

// Debug prints a message at the debug level.
func (l *Logger) Debug(message string) {
	fmt.Fprintf(l.Output, "::debug::%s\n", message)
}

// Info prints a message at the info level.
func (l *Logger) Info(message string) {
	fmt.Fprintln(l.Output, message)
}

// Warn prints a message at the warning level.
func (l *Logger) Warn(message string) {
	fmt.Fprintf(l.Output, "::warning::%s\n", message)
}

// Error prints a message at the error level.
func (l *Logger) Error(message string) {
	fmt.Fprintf(l.Output, "::error::%s\n", message)
}

// SetSecret tells the actions environment to mask the supplied value.
func (l *Logger) SetSecret(value string) {
	fmt.Fprintf(l.Output, "::add-mask::%s\n", value)
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
