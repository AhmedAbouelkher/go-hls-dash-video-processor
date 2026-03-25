package main

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestFFmpegGenError_Error(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		innerErr error
		output   string
		want     string
	}{
		{
			name:     "exit code error",
			code:     1,
			innerErr: nil,
			output:   "ffmpeg error output",
			want:     "1 - ffmpeg error output",
		},
		{
			name:     "inner error",
			code:     -1,
			innerErr: errors.New("command not found"),
			output:   "",
			want:     "command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &FFmpegGenError{
				Code:     tt.code,
				innerErr: tt.innerErr,
				Output:   tt.output,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFFmpegGenError_Log(t *testing.T) {
	e := &FFmpegGenError{
		Code:   1,
		Output: "test output",
		Stack:  "test stack",
	}
	fields := e.Log()
	if fields["code"] != 1 {
		t.Errorf("Log() code = %v, want 1", fields["code"])
	}
	if fields["output"] != "test output" {
		t.Errorf("Log() output = %v, want 'test output'", fields["output"])
	}
	if fields["stack"] != "test stack" {
		t.Errorf("Log() stack = %v, want 'test stack'", fields["stack"])
	}
}

func TestNewFFmpegGenError(t *testing.T) {
	t.Run("with exit error", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 42")
		err := cmd.Run()
		output := "command failed"
		
		ffmpegErr := NewFFmpegGenError(err, output)
		
		if ffmpegErr.Code != 42 {
			t.Errorf("NewFFmpegGenError() code = %v, want 42", ffmpegErr.Code)
		}
		if ffmpegErr.Output != output {
			t.Errorf("NewFFmpegGenError() output = %v, want %v", ffmpegErr.Output, output)
		}
		if ffmpegErr.Stack == "" {
			t.Error("NewFFmpegGenError() stack should not be empty")
		}
	})

	t.Run("with generic error", func(t *testing.T) {
		err := errors.New("generic error")
		output := "some output"
		
		ffmpegErr := NewFFmpegGenError(err, output)
		
		if ffmpegErr.Code != -1 {
			t.Errorf("NewFFmpegGenError() code = %v, want -1", ffmpegErr.Code)
		}
		if ffmpegErr.innerErr != err {
			t.Errorf("NewFFmpegGenError() innerErr = %v, want %v", ffmpegErr.innerErr, err)
		}
	})
}

func TestTrimStack(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		check func(string) bool
	}{
		{
			name:  "short stack",
			input: []byte("line1\nline2\nline3"),
			check: func(s string) bool {
				return strings.Contains(s, "line1") && strings.Contains(s, "line2") && strings.Contains(s, "line3")
			},
		},
		{
			name:  "long stack",
			input: []byte(strings.Repeat("line\n", 50)),
			check: func(s string) bool {
				lines := strings.Split(s, "\n")
				return len(lines) <= 36
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimStack(tt.input)
			if !tt.check(got) {
				t.Errorf("trimStack() failed check for %s", tt.name)
			}
		})
	}
}
