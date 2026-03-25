package main

import (
	"testing"
	"time"
)

func TestIsValidResolution(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid 240p",
			input: Res240p,
			want:  true,
		},
		{
			name:  "valid 360p",
			input: Res360p,
			want:  true,
		},
		{
			name:  "valid 480p",
			input: Res480p,
			want:  true,
		},
		{
			name:  "valid 720p",
			input: Res720p,
			want:  true,
		},
		{
			name:  "valid 1080p",
			input: Res1080p,
			want:  true,
		},
		{
			name:  "invalid resolution",
			input: "1440p",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidResolution(tt.input); got != tt.want {
				t.Errorf("isValidResolution(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolution_String(t *testing.T) {
	r := &resolution{Width: 1280, Height: 720}
	want := "1280:720"
	if got := r.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestResolution_StringV2(t *testing.T) {
	r := &resolution{Width: 1920, Height: 1080}
	want := "1920x1080"
	if got := r.StringV2(); got != want {
		t.Errorf("StringV2() = %v, want %v", got, want)
	}
}

func TestResolution_Name(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   string
	}{
		{
			name:   "240p",
			height: 240,
			want:   "240p",
		},
		{
			name:   "720p",
			height: 720,
			want:   "720p",
		},
		{
			name:   "1080p",
			height: 1080,
			want:   "1080p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolution{Height: tt.height}
			if got := r.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolutionsMap(t *testing.T) {
	m := resolutionsMap()
	
	if len(m) != 5 {
		t.Errorf("resolutionsMap() length = %v, want 5", len(m))
	}
	
	if _, ok := m[Res240p]; !ok {
		t.Error("resolutionsMap() missing 240p")
	}
	if _, ok := m[Res720p]; !ok {
		t.Error("resolutionsMap() missing 720p")
	}
}

func TestBytesToGB(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  float64
	}{
		{
			name:  "1 GB",
			bytes: 1000000000,
			want:  1.0,
		},
		{
			name:  "0 bytes",
			bytes: 0,
			want:  0.0,
		},
		{
			name:  "500 MB",
			bytes: 500000000,
			want:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesToGB(tt.bytes); got != tt.want {
				t.Errorf("bytesToGB(%v) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestSecondsToHours(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		want    float64
	}{
		{
			name:    "1 hour",
			seconds: 3600,
			want:    1.0,
		},
		{
			name:    "30 minutes",
			seconds: 1800,
			want:    0.5,
		},
		{
			name:    "0 seconds",
			seconds: 0,
			want:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := secondsToHours(tt.seconds); got != tt.want {
				t.Errorf("secondsToHours(%v) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestLgZero(t *testing.T) {
	tests := []struct {
		name string
		v    int
		dft  int
		want int
	}{
		{
			name: "zero value returns default",
			v:    0,
			dft:  100,
			want: 100,
		},
		{
			name: "non-zero value returns value",
			v:    50,
			dft:  100,
			want: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lgZero(tt.v, tt.dft); got != tt.want {
				t.Errorf("lgZero(%v, %v) = %v, want %v", tt.v, tt.dft, got, tt.want)
			}
		})
	}
}

func TestResolution_CalcTimeout(t *testing.T) {
	r := &resolution{
		Height:  720,
		timeout: 30 * time.Minute,
	}

	tests := []struct {
		name     string
		calc     *TimeoutCalc
		minTime  time.Duration
	}{
		{
			name:     "small file",
			calc:     &TimeoutCalc{size: 500000000, duration: 60},
			minTime:  30 * time.Minute,
		},
		{
			name:     "large file",
			calc:     &TimeoutCalc{size: 5000000000, duration: 3600},
			minTime:  30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.CalcTimeout(tt.calc)
			if got < tt.minTime {
				t.Errorf("CalcTimeout() = %v, want at least %v", got, tt.minTime)
			}
		})
	}
}

func TestGenerationResolutions_Resolutions(t *testing.T) {
	res1 := &resolution{Height: 240}
	res2 := &resolution{Height: 720}
	
	g := GenerationResolutions{
		&GenerateResOutput{Res: res1},
		&GenerateResOutput{Res: res2},
	}

	got := g.Resolutions()
	if len(got) != 2 {
		t.Errorf("Resolutions() length = %v, want 2", len(got))
	}
	if got[0] != "240p" {
		t.Errorf("Resolutions()[0] = %v, want 240p", got[0])
	}
	if got[1] != "720p" {
		t.Errorf("Resolutions()[1] = %v, want 720p", got[1])
	}
}
