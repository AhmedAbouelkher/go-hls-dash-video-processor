package main

import (
	"fmt"
	"maps"
	"os"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[1;34m"
	colorCyan   = "\033[1;36m"
	colorYellow = "\033[1;33m"
	colorRed    = "\033[1;31m"
	colorPurple = "\033[1;35m"
	colorGreen  = "\033[1;32m"
)

// Fields is a map of key-value pairs for structured logging.
type Fields map[string]any

type Logger struct {
	fields Fields
	err    error
}

func NewLogger() *Logger { return &Logger{fields: Fields{}} }

func (l *Logger) WithFields(f Fields) *Logger {
	merged := make(Fields, len(l.fields)+len(f))
	maps.Copy(merged, l.fields)
	maps.Copy(merged, f)
	return &Logger{fields: merged, err: l.err}
}

func (l *Logger) WithField(k string, v any) *Logger { return l.WithFields(Fields{k: v}) }
func (l *Logger) WithError(err error) *Logger       { return &Logger{fields: l.fields, err: err} }
func (l *Logger) Info(args ...any)                  { l.log("INFO", fmt.Sprint(args...)) }
func (l *Logger) Infof(f string, a ...any)          { l.log("INFO", fmt.Sprintf(f, a...)) }
func (l *Logger) Debug(args ...any)                 { l.log("DEBUG", fmt.Sprint(args...)) }
func (l *Logger) Debugf(f string, a ...any)         { l.log("DEBUG", fmt.Sprintf(f, a...)) }
func (l *Logger) Warning(args ...any)               { l.log("WARN", fmt.Sprint(args...)) }
func (l *Logger) Error(args ...any)                 { l.log("ERROR", fmt.Sprint(args...)) }
func (l *Logger) Fatal(args ...any) {
	l.log("FATAL", fmt.Sprint(args...))
	os.Exit(1)
}

func (l *Logger) log(level, msg string) {
	color := colorReset
	switch level {
	case "INFO":
		color = colorBlue
	case "DEBUG":
		color = colorCyan
	case "WARN":
		color = colorYellow
	case "ERROR":
		color = colorRed
	case "FATAL":
		color = colorPurple
	}

	var line strings.Builder
	fmt.Fprintf(&line, "%s %s[%s]%s %s", time.Now().Format("2006/01/02 15:04:05"), color, level, colorReset, msg)
	if l.err != nil {
		fmt.Fprintf(&line, " %serror%s=%q", colorGreen, colorReset, l.err.Error())
	}
	for k, v := range l.fields {
		fmt.Fprintf(&line, " %s%s%s=%v", colorGreen, k, colorReset, v)
	}
	fmt.Fprintln(os.Stderr, line.String())
}
