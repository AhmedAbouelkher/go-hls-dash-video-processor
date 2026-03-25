package main

import (
	"testing"

	"github.com/AhmedAbouelkher/go-hls-dash-video-processor/gabs"
)

func TestJInt(t *testing.T) {
	json := `{"value": 42}`
	container, err := gabs.ParseJSON([]byte(json))
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	got := jInt(container.S("value"))
	want := 42
	if got != want {
		t.Errorf("jInt() = %v, want %v", got, want)
	}
}

func TestJStr(t *testing.T) {
	json := `{"name": "test"}`
	container, err := gabs.ParseJSON([]byte(json))
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	got := jStr(container.S("name"))
	want := "test"
	if got != want {
		t.Errorf("jStr() = %v, want %v", got, want)
	}
}

func TestMustParseFloat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float32
	}{
		{
			name:  "valid float",
			input: "3.14",
			want:  3.14,
		},
		{
			name:  "valid integer",
			input: "42",
			want:  42.0,
		},
		{
			name:  "invalid string",
			input: "invalid",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseFloat(tt.input); got != tt.want {
				t.Errorf("mustParseFloat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMustParseInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "valid integer",
			input: "123",
			want:  123,
		},
		{
			name:  "zero",
			input: "0",
			want:  0,
		},
		{
			name:  "invalid string",
			input: "abc",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseInt(tt.input); got != tt.want {
				t.Errorf("mustParseInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMustParseDiv(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float32
	}{
		{
			name:  "valid division",
			input: "30/1",
			want:  30.0,
		},
		{
			name:  "fractional result",
			input: "25/2",
			want:  12.5,
		},
		{
			name:  "invalid format",
			input: "30",
			want:  0,
		},
		{
			name:  "invalid numerator",
			input: "abc/2",
			want:  0,
		},
		{
			name:  "invalid denominator",
			input: "30/abc",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseDiv(tt.input); got != tt.want {
				t.Errorf("mustParseDiv(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDurTms(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		want     int
	}{
		{
			name:     "100 seconds",
			duration: 100,
			want:     20,
		},
		{
			name:     "0 seconds",
			duration: 0,
			want:     0,
		},
		{
			name:     "50 seconds",
			duration: 50,
			want:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := durTms(tt.duration); got != tt.want {
				t.Errorf("durTms(%v) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatTimeSeconds(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{
			name:    "1 hour",
			seconds: 3600,
			want:    "01:00:00",
		},
		{
			name:    "1 minute 30 seconds",
			seconds: 90,
			want:    "00:01:30",
		},
		{
			name:    "0 seconds",
			seconds: 0,
			want:    "00:00:00",
		},
		{
			name:    "complex time",
			seconds: 3661,
			want:    "01:01:01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTimeSeconds(tt.seconds); got != tt.want {
				t.Errorf("formatTimeSeconds(%v) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestParseRes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantW   int
		wantH   int
		wantErr bool
	}{
		{
			name:    "valid resolution",
			input:   "1920x1080",
			wantW:   1920,
			wantH:   1080,
			wantErr: false,
		},
		{
			name:    "small resolution",
			input:   "640x480",
			wantW:   640,
			wantH:   480,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "1920",
			wantW:   0,
			wantH:   0,
			wantErr: true,
		},
		{
			name:    "invalid width",
			input:   "abcx480",
			wantW:   0,
			wantH:   0,
			wantErr: true,
		},
		{
			name:    "invalid height",
			input:   "1920xabc",
			wantW:   0,
			wantH:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH, err := parseRes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRes(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if gotW != tt.wantW {
				t.Errorf("parseRes(%q) width = %v, want %v", tt.input, gotW, tt.wantW)
			}
			if gotH != tt.wantH {
				t.Errorf("parseRes(%q) height = %v, want %v", tt.input, gotH, tt.wantH)
			}
		})
	}
}

func TestVideoResolution_String(t *testing.T) {
	v := &VideoResolution{Width: 1280, Height: 720}
	want := "1280x720"
	if got := v.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}
