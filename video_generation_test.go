package main

import (
	"errors"
	"os"
	"testing"
)

func TestVideoGenerateInput_LogDetails(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_video_*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	input := &VideoGenerateInput{
		Input:  tmpFile,
		Output: "/output/path",
		Type:   HLSProcessType,
	}

	fields := input.LogDetails()

	if fields["input"] != tmpFile.Name() {
		t.Errorf("LogDetails input = %v, want %v", fields["input"], tmpFile.Name())
	}
	if fields["output"] != "/output/path" {
		t.Errorf("LogDetails output = %v, want /output/path", fields["output"])
	}
	if fields["type"] != HLSProcessType {
		t.Errorf("LogDetails type = %v, want %v", fields["type"], HLSProcessType)
	}
}

func TestBuildGenErrLogDetails(t *testing.T) {
	t.Run("with FFmpegGenError", func(t *testing.T) {
		err := &FFmpegGenError{
			Code:   1,
			Output: "test output",
			Stack:  "test stack",
		}
		fields := buildGenErrLogDetails(err)
		
		if fields["code"] != 1 {
			t.Errorf("buildGenErrLogDetails code = %v, want 1", fields["code"])
		}
		if fields["output"] != "test output" {
			t.Errorf("buildGenErrLogDetails output = %v, want 'test output'", fields["output"])
		}
	})

	t.Run("with generic error", func(t *testing.T) {
		err := errors.New("generic error")
		fields := buildGenErrLogDetails(err)
		
		if len(fields) != 0 {
			t.Errorf("buildGenErrLogDetails with generic error should return empty fields, got %v", fields)
		}
	})
}

func TestResolutionsSettings_LogDetails(t *testing.T) {
	settings := &ResolutionsSettings{
		DisableMP4Fallback: true,
		EnabledResolutions: []string{"720p", "1080p"},
		Bitrate240p:        400,
		Bitrate360p:        800,
		Bitrate480p:        1000,
		Bitrate720p:        1500,
		Bitrate1080p:       3500,
	}

	fields := settings.LogDetails()

	if fields["disable_mp4_fallback"] != true {
		t.Errorf("LogDetails disable_mp4_fallback = %v, want true", fields["disable_mp4_fallback"])
	}
	if fields["enabled_resolutions"] != "720p,1080p" {
		t.Errorf("LogDetails enabled_resolutions = %v, want '720p,1080p'", fields["enabled_resolutions"])
	}
	if fields["bitrate_240p"] != 400 {
		t.Errorf("LogDetails bitrate_240p = %v, want 400", fields["bitrate_240p"])
	}
	if fields["bitrate_720p"] != 1500 {
		t.Errorf("LogDetails bitrate_720p = %v, want 1500", fields["bitrate_720p"])
	}
}
