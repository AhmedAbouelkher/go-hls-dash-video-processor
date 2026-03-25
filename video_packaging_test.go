package main

import "testing"

func TestIsValidProcessType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid hls",
			input: "hls",
			want:  true,
		},
		{
			name:  "valid hls_and_dash",
			input: "hls_and_dash",
			want:  true,
		},
		{
			name:  "invalid type",
			input: "invalid",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "dash only",
			input: "dash",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidProcessType(tt.input); got != tt.want {
				t.Errorf("isValidProcessType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateHlsInput_hasAudio(t *testing.T) {
	audioPath := "path/to/audio.aac"
	tests := []struct {
		name  string
		input *GenerateHlsInput
		want  bool
	}{
		{
			name:  "has audio",
			input: &GenerateHlsInput{Audio: &audioPath},
			want:  true,
		},
		{
			name:  "no audio",
			input: &GenerateHlsInput{Audio: nil},
			want:  false,
		},
		{
			name: "empty audio path",
			input: &GenerateHlsInput{Audio: func() *string {
				s := ""
				return &s
			}()},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.hasAudio(); got != tt.want {
				t.Errorf("hasAudio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateHlsInput_audioPath(t *testing.T) {
	audioPath := "path/to/audio.aac"
	emptyPath := ""

	tests := []struct {
		name  string
		input *GenerateHlsInput
		want  string
	}{
		{
			name:  "has audio path",
			input: &GenerateHlsInput{Audio: &audioPath},
			want:  "path/to/audio.aac",
		},
		{
			name:  "nil audio",
			input: &GenerateHlsInput{Audio: nil},
			want:  "",
		},
		{
			name:  "empty audio path",
			input: &GenerateHlsInput{Audio: &emptyPath},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.audioPath(); got != tt.want {
				t.Errorf("audioPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
