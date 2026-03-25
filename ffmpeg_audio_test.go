package main

import (
	"reflect"
	"testing"
)

func TestMustIntVal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "valid integer",
			input: "12345",
			want:  12345,
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
			if got := mustIntVal(tt.input); got != tt.want {
				t.Errorf("mustIntVal(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "valid output",
			input: "bit_rate=128000\nsample_rate=44100\nchannels=2",
			want: map[string]string{
				"bit_rate":    "128000",
				"sample_rate": "44100",
				"channels":    "2",
			},
		},
		{
			name:  "single line",
			input: "bit_rate=96000",
			want: map[string]string{
				"bit_rate": "96000",
			},
		},
		{
			name:  "empty string",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "invalid line without equals",
			input: "bit_rate=128000\ninvalid_line\nchannels=2",
			want: map[string]string{
				"bit_rate": "128000",
				"channels": "2",
			},
		},
		{
			name:  "line with multiple equals",
			input: "key=value=extra",
			want:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOutput(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseOutput(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewAudioData(t *testing.T) {
	m := map[string]string{
		"bit_rate":       "128000",
		"sample_rate":    "44100",
		"channel_layout": "stereo",
		"channels":       "2",
	}
	output := "test_output"

	ad := newAudioData(m, output)

	if ad.BitRate != 128 {
		t.Errorf("BitRate = %v, want 128", ad.BitRate)
	}
	if ad.SampleRate != 44100 {
		t.Errorf("SampleRate = %v, want 44100", ad.SampleRate)
	}
	if ad.ChannelLayout != "stereo" {
		t.Errorf("ChannelLayout = %v, want stereo", ad.ChannelLayout)
	}
	if ad.Channels != 2 {
		t.Errorf("Channels = %v, want 2", ad.Channels)
	}
	if ad.Output != output {
		t.Errorf("Output = %v, want %v", ad.Output, output)
	}
}

func TestAudioData_LogDetails(t *testing.T) {
	ad := &AudioData{
		BitRate:       128,
		SampleRate:    44100,
		ChannelLayout: "stereo",
		Channels:      2,
		Output:        "test.aac",
	}

	fields := ad.LogDetails()

	if fields["bit_rate"] != 128 {
		t.Errorf("LogDetails bit_rate = %v, want 128", fields["bit_rate"])
	}
	if fields["sample_rate"] != 44100 {
		t.Errorf("LogDetails sample_rate = %v, want 44100", fields["sample_rate"])
	}
	if fields["channel_layout"] != "stereo" {
		t.Errorf("LogDetails channel_layout = %v, want stereo", fields["channel_layout"])
	}
	if fields["channels"] != 2 {
		t.Errorf("LogDetails channels = %v, want 2", fields["channels"])
	}
	if fields["output"] != "test.aac" {
		t.Errorf("LogDetails output = %v, want test.aac", fields["output"])
	}
}
