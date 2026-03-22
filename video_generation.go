package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
)

type VideoGenerateInput struct {
	Ctx      context.Context
	VideoID  string
	Type     ProcessType
	Settings *ResolutionsSettings
	Input    *os.File
	Output   string
}

func (g *VideoGenerateInput) LogDetails() Fields {
	return Fields{"input": g.Input.Name(), "output": g.Output, "type": g.Type}
}

type VideoMetadata struct {
	Size       int64            `json:"size"`
	Resolution *VideoResolution `json:"resolution"`
	Duration   int64            `json:"duration"`
	SourceSize int64            `json:"source_size"`
}
type VideoGenerateOutput struct {
	Target          string                `json:"target"`
	VideoMetadata   *VideoMetadata        `json:"video_metadata"`
	Thumbnail       string                `json:"thumbnail"`
	VideoGeneration GenerationResolutions `json:"video_generation"`
	AudioOutput     *ACCAudioOutput       `json:"audio_output"`
	VideoPackaging  *PackagingOutput      `json:"video_packaging"`
}

func buildGenErrLogDetails(err error) Fields {
	if e, ok := err.(*FFmpegGenError); ok {
		return e.Log()
	}
	return Fields{}
}

func GenerateAndPackageVideo(in *VideoGenerateInput) (*VideoGenerateOutput, error) {
	inputFileName := in.Input.Name()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(in.Ctx)
	defer cancel()

	if err := CheckIntegrity(ctx, inputFileName); err != nil {
		logger.WithError(err).WithFields(Fields{
			"input":  in.Input.Name(),
			"output": in.Output,
			"type":   in.Type,
		}).Error("💩 Video file is corrupted")
		return nil, err
	} else {
		logger.WithFields(in.LogDetails()).Info("✅ video file is valid")
	}

	srcSize, err := GetVideoSize(ctx, inputFileName)
	if err != nil {
		logger.WithError(err).WithFields(in.LogDetails()).Error("failed to get video size")
	}

	duration, err := GetVideoDuration(ctx, inputFileName)
	if err != nil {
		logger.WithError(err).WithFields(in.LogDetails()).Error("failed to get video duration")
	}

	wg.Add(2)
	ac := make(chan audioResult, 1)
	go func() {
		defer wg.Done()
		defer close(ac)

		generateAudio(ctx, ac, in.VideoID, inputFileName, in.Output)
	}()

	vc := make(chan videoResult, 1)
	go func() {
		defer wg.Done()
		defer close(vc)

		i := &videoGenInput{
			videoID:  in.VideoID,
			vc:       vc,
			settings: in.Settings,
			input:    inputFileName,
			target:   in.Output,
			size:     srcSize,
			Duration: duration,
		}
		generateVideo(ctx, i)
	}()

	wg.Wait()

	a := <-ac
	if err := a.err; err != nil {
		logger.WithError(err).
			WithFields(in.LogDetails()).
			WithFields(buildGenErrLogDetails(err)).
			Error("failed to generate audio")
		return nil, err
	}

	v := <-vc
	if err := v.err; err != nil {
		logger.WithError(err).WithFields(in.LogDetails()).Error("failed to generate video")
		return nil, err
	}

	cin := &videoComposeInput{
		videoID: in.VideoID,
		vr:      &v,
		ar:      &a,
		typ:     in.Type,
		target:  in.Output,
	}
	pout, err := packageVideo(ctx, cin)
	if err != nil {
		if pout == nil {
			pout = &PackagingOutput{}
		}
		logger.WithError(err).WithFields(in.LogDetails()).WithFields(buildGenErrLogDetails(err)).Error("failed to package video")
		return nil, err
	}
	thInput := &ThumbnailInput{
		Duration: duration,
		Input:    inputFileName,
		Target:   in.Output,
	}
	thumbnail, err := GenThumbnail(ctx, thInput)
	if err != nil {
		logger.WithError(err).WithFields(in.LogDetails()).WithFields(Fields{
			"output": thumbnail,
		}).Error("failed to generate video thumbnail")
		thumbnail = "" // reset thumbnail
	}

	if err := cleanFiles(v.resolutions, in.Settings); err != nil {
		logger.WithError(err).WithFields(in.LogDetails()).Error("failed to clean up files")
	}

	ts, err := GetDirSize(ctx, in.Output)
	if err != nil {
		logger.WithError(err).WithFields(in.LogDetails()).WithFields(Fields{
			"output": ts,
		}).Error("failed to get processed video size")
		ts = 0 // reset size
	}

	return &VideoGenerateOutput{
		Target: in.Output,
		VideoMetadata: &VideoMetadata{
			Size:       ts,
			Resolution: v.vRes,
			SourceSize: srcSize,
			Duration:   duration,
		},
		Thumbnail:       filepath.Base(thumbnail),
		VideoGeneration: v.resolutions,
		AudioOutput:     a.af,
		VideoPackaging:  pout,
	}, nil
}

func cleanFiles(genRes GenerationResolutions, settings *ResolutionsSettings) error {
	wAudioFiles := []string{}
	noAudioFiles := []string{}
	for _, r := range genRes {
		noAudioFiles = append(noAudioFiles, r.VNoAudio)
		wAudioFiles = append(wAudioFiles, r.Video)
	}
	for _, r := range noAudioFiles {
		if err := os.Remove(r); err != nil {
			return err
		}
	}
	if settings.DisableMP4Fallback {
		for _, r := range wAudioFiles {
			if err := os.Remove(r); err != nil {
				return err
			}
		}
	}
	return nil
}
