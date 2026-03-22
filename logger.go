package main

import (
	"maps"
	"fmt"
	"os"
	"strings"
	"time"
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
func (l *Logger) Warning(args ...any)               { l.log("WARN", fmt.Sprint(args...)) }
func (l *Logger) Error(args ...any)                 { l.log("ERROR", fmt.Sprint(args...)) }
func (l *Logger) Fatal(args ...any) {
	l.log("FATAL", fmt.Sprint(args...))
	os.Exit(1)
}

func (l *Logger) log(level, msg string) {
	var line strings.Builder
	line.WriteString(fmt.Sprintf("%s [%s] %s", time.Now().Format("2006/01/02 15:04:05"), level, msg))
	if l.err != nil {
		line.WriteString(fmt.Sprintf(" error=%q", l.err.Error()))
	}
	for k, v := range l.fields {
		line.WriteString(fmt.Sprintf(" %s=%v", k, v))
	}
	fmt.Fprintln(os.Stderr, line.String())
}
