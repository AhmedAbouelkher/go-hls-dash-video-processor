package main

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type audioResult struct {
	af  *ACCAudioOutput
	err error
}

func generateAudio(ctx context.Context, ac chan audioResult, videoID, i, tf string) {
	in := &GenerateAudioInput{
		Input:  i,
		Target: tf,
	}

	logger.WithFields(Fields{
		"video_id": videoID,
		"input":    in.Input,
		"target":   in.Target,
	}).Info("🔉 generating audio")

	// Generate [audio] using ffmpeg
	af, err := GenerateACC(ctx, in.Input, in.Target)
	if err != nil {
		ac <- audioResult{err: err}
		return
	}
	if af == nil {
		logger.WithFields(Fields{
			"input":    in.Input,
			"target":   in.Target,
			"video_id": videoID,
		}).Warning("no audio data found")
	}
	ac <- audioResult{af, nil}
}

// MARK:- ACC

type GenerateAudioInput struct {
	Input  string
	Target string
}

type ACCAudioOutput struct {
	Path string `json:"path"`

	GenerationDuration time.Duration `json:"generation_duration"`
	AData              *AudioData    `json:"audio_data"`
}

func GenerateACC(ctx context.Context, input string, target string) (*ACCAudioOutput, error) {
	tCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	data, err := getAudioMetadata(ctx, input)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	output := fmt.Sprintf("%s/play_audio.aac", target)
	c := fmt.Sprintf("ffmpeg -v error -y -i %s -vn -acodec copy -vn %s", input, output)
	// check if input is mkv
	if strings.HasSuffix(input, ".mkv") {
		c = fmt.Sprintf("ffmpeg -v error -y -i %s -map 0:a -c:a aac %s", input, output)
	}
	args := strings.Split(c, " ")
	t := time.Now()
	co, err := exec.CommandContext(tCtx, args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(co))
	}
	return &ACCAudioOutput{
		Path:               output,
		GenerationDuration: time.Since(t),
		AData:              data,
	}, err
}

// MARK:- Utilities

type AudioData struct {
	BitRate       int
	SampleRate    int
	ChannelLayout string
	Channels      int

	Output string
}

func newAudioData(m map[string]string, output string) *AudioData {
	return &AudioData{
		BitRate:       mustIntVal(m["bit_rate"]) / 1000,
		SampleRate:    mustIntVal(m["sample_rate"]),
		ChannelLayout: m["channel_layout"],
		Channels:      mustIntVal(m["channels"]),

		Output: output,
	}
}

func mustIntVal(v string) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return i
}

func (a *AudioData) LogDetails() Fields {
	return Fields{
		"bit_rate":       a.BitRate,
		"sample_rate":    a.SampleRate,
		"channel_layout": a.ChannelLayout,
		"channels":       a.Channels,
		"output":         a.Output,
	}
}

func getAudioMetadata(ctx context.Context, i string) (*AudioData, error) {
	tCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	c := "ffprobe -v error -select_streams a:0 -show_entries stream=bit_rate,sample_rate,channel_layout,channels -of default=noprint_wrappers=1"
	c += " " + fmt.Sprintf("'%s'", i)
	out, err := exec.CommandContext(tCtx, "bash", "-c", c).CombinedOutput()
	d := string(out)
	if err != nil {
		return nil, NewFFmpegGenError(err, d)
	}
	if d == "" {
		return nil, nil
	}
	o := strings.Trim(d, "\n")
	m := parseOutput(o)
	return newAudioData(m, o), nil
}

func parseOutput(o string) map[string]string {
	m := make(map[string]string)
	for _, v := range strings.Split(o, "\n") {
		kv := strings.Split(v, "=")
		// skip empty lines, and lines without a key and value
		if len(kv) != 2 {
			continue
		}
		m[kv[0]] = kv[1]
	}
	return m
}
