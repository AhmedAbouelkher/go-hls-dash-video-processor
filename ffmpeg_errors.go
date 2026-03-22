package main

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"strings"
)

type FFmpegGenError struct {
	Code     int
	innerErr error
	Output   string
	Stack    string
}

func (e *FFmpegGenError) Error() string {
	if e.Code == -1 {
		return e.innerErr.Error()
	}
	return fmt.Sprintf("%d - %s", e.Code, e.Output)
}

func (e *FFmpegGenError) Log() Fields {
	return Fields{
		"code":   e.Code,
		"output": e.Output,
		"stack":  e.Stack,
	}
}

func NewFFmpegGenError(err error, output string) *FFmpegGenError {
	stack := trimStack(debug.Stack())
	e := &FFmpegGenError{
		innerErr: err,
		Output:   output,
		Stack:    stack,
	}
	if eerr, ok := err.(*exec.ExitError); ok {
		e.Code = eerr.ExitCode()
	} else {
		e.Code = -1
	}
	return e
}

func trimStack(s []byte) string {
	stackList := strings.Split(string(s), "\n")
	var stack strings.Builder
	if len(stackList) > 35 {
		stackList = stackList[:35]
	}
	for _, v := range stackList {
		stack.WriteString(v + "\n")
	}
	return stack.String()
}
