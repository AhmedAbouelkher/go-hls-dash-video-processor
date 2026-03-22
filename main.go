package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	sourceVideo, destination string
	force                    bool

	rawOutputType string
	outputType    ProcessType

	rawResolutions     string
	enabledResolutions []string

	logger = NewLogger()
)

func main() {
	if len(os.Args) < 2 {
		logger.Fatal("source video is required")
	}
	sourceVideo = os.Args[1]

	flag.StringVar(&destination, "d", "", "The path to the destination directory")
	flag.BoolVar(&force, "f", false, "Force the generation of the video")
	flag.StringVar(&rawOutputType, "t", "hls", "The type of output to generate")
	flag.StringVar(&rawResolutions, "r", "", "The resolutions to generate as a comma separated list")
	flag.CommandLine.Parse(os.Args[2:])

	// parse raw resolutions
	if rawResolutions != "" {
		parts := strings.Split(rawResolutions, ",")
		enabledResolutions = make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				if !isValidResolution(trimmed) {
					logger.WithField("resolution", trimmed).Fatal("invalid resolution")
				}
				enabledResolutions = append(enabledResolutions, trimmed)
			}
		}
	} else {
		enabledResolutions = []string{Res240p, Res360p, Res480p, Res720p, Res1080p}
	}

	if rawOutputType != "" {
		if !isValidProcessType(rawOutputType) {
			logger.WithField("output_type", rawOutputType).Fatal("invalid output type")
		}
		outputType = ProcessType(rawOutputType)
	} else {
		outputType = HLSProcessType
	}

	// open source video
	input, err := os.Open(sourceVideo)
	if err != nil {
		logger.WithError(err).Fatal("failed to open source video")
	}
	defer input.Close()

	// if destination is not set, use the base name of the source video
	if destination == "" {
		dir := strings.ToLower(strings.ReplaceAll(strings.Split(filepath.Base(input.Name()), ".")[0], " ", "_"))
		destination = dir
	}

	// check if directory exists, if exists, throw error
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		if force {
			if err := os.RemoveAll(destination); err != nil {
				logger.WithError(err).Fatal("failed to remove existing destination directory")
			}
		} else {
			logger.WithField("directory", destination).Fatal("destination directory already exists")
		}
	}

	if err := os.MkdirAll(destination, 0755); err != nil {
		logger.WithError(err).Fatal("failed to create destination directory")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	videoGenerateInput := &VideoGenerateInput{
		Ctx:     ctx,
		VideoID: sourceVideo,
		Type:    outputType,
		Settings: &ResolutionsSettings{
			EnabledResolutions: enabledResolutions,
		},
		Input:  input,
		Output: destination,
	}
	videoGenerateOutput, err := GenerateAndPackageVideo(videoGenerateInput)
	if err != nil {
		logger.WithError(err).Fatal("failed to generate video")
	}

	generatedRes := []string{}
	for _, r := range videoGenerateOutput.VideoGeneration {
		d := fmt.Sprintf("%s: %s, %s", r.Res.Name(), r.Video, r.VNoAudio)
		generatedRes = append(generatedRes, d)
	}

	logger.WithFields(videoGenerateOutput.VideoGeneration.LogDetails()).
		WithField("thumbnail", videoGenerateOutput.Thumbnail).
		WithField("generated_resolutions", strings.Join(generatedRes, ", ")).
		Info("🎉 video generated successfully")
}
