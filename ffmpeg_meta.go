package main

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AhmedAbouelkher/go-hls-dash-video-processor/gabs"
)

type VideoFormate string

var (
	Duration = VideoFormate("duration")
	BitRate  = VideoFormate("bit_rate")
	Size     = VideoFormate("size")
)

func CheckIntegrity(ctx context.Context, input string) error {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	c := fmt.Sprintf("ffmpeg -v error -i '%s' -c copy -f null - 2>&1", input)
	o, err := exec.CommandContext(tCtx, "bash", "-c", c).CombinedOutput()
	if err != nil {
		return NewFFmpegGenError(err, string(o))
	}
	return nil
}

type VideoInfo struct {
	Hight int `json:"hight" bson:"hight"`
	Width int `json:"width" bson:"width"`

	BitRate  int     `json:"bit_rate" bson:"bit_rate"`
	FPS      float32 `json:"r_frame_rate" bson:"fps"`
	Codec    string  `json:"codec_name" bson:"codec"`
	Duration float32 `json:"duration" bson:"duration"`
	Profile  string  `json:"profile" bson:"profile"`
}

type AudioInfo struct {
	Codec         string  `json:"codec_name" bson:"codec"`
	BitRate       int     `json:"bit_rate" bson:"bit_rate"`
	SampleRate    int     `json:"sample_rate" bson:"sample_rate"`
	Channels      int     `json:"channels" bson:"channels"`
	Profile       string  `json:"profile" bson:"profile"`
	ChannelLayout string  `json:"channel_layout" bson:"channel_layout"`
	Duration      float32 `json:"duration" bson:"duration"`
}

func GetVideoInfo(ctx context.Context, i string) (*VideoInfo, error) {
	s := "width,height,bit_rate,codec_name,r_frame_rate,duration,profile"
	c := fmt.Sprintf("ffprobe -v error -select_streams v:0 -show_entries stream=%s -of json %s", s, i)

	o, err := exec.CommandContext(ctx, "bash", "-c", c).CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(o))
	}

	return parseVOut(o)
}

func parseVOut(o []byte) (*VideoInfo, error) {
	jp, err := gabs.ParseJSON(o)
	if err != nil {
		return nil, err
	}
	str := jp.S("streams").Children()
	if len(str) == 0 {
		return nil, nil
	}
	vst := str[0]

	vi := &VideoInfo{
		Width:    jInt(vst.S("width")),
		Hight:    jInt(vst.S("height")),
		BitRate:  mustParseInt(jStr(vst.S("bit_rate"))),
		FPS:      mustParseDiv(jStr(vst.S("r_frame_rate"))),
		Codec:    jStr(vst.S("codec_name")),
		Duration: mustParseFloat(jStr(vst.S("duration"))),
		Profile:  jStr(vst.S("profile")),
	}

	return vi, nil
}
func GetAudioInfo(ctx context.Context, i string) (*AudioInfo, error) {
	s := "codec_name,bit_rate,sample_rate,channels,profile,channel_layout,duration"
	c := fmt.Sprintf("ffprobe -v error -select_streams a:0 -show_entries stream=%s -of json %s", s, i)
	o, err := exec.CommandContext(ctx, "bash", "-c", c).CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(o))
	}
	return parseAOut(o)
}

func parseAOut(o []byte) (*AudioInfo, error) {
	jp, err := gabs.ParseJSON(o)
	if err != nil {
		return nil, err
	}
	str := jp.S("streams").Children()
	if len(str) == 0 {
		return nil, nil
	}
	vst := str[0]

	ai := &AudioInfo{
		Codec:         jStr(vst.S("codec_name")),
		BitRate:       mustParseInt(jStr(vst.S("bit_rate"))),
		SampleRate:    mustParseInt(jStr(vst.S("sample_rate"))),
		Channels:      mustParseInt(jStr(vst.S("channels"))),
		Profile:       jStr(vst.S("profile")),
		ChannelLayout: jStr(vst.S("channel_layout")),
		Duration:      mustParseFloat(jStr(vst.S("duration"))),
	}

	return ai, nil
}

func jInt(c *gabs.Container) int { return int(c.Data().(float64)) }

func jStr(c *gabs.Container) string { return fmt.Sprintf("%v", c.Data()) }

func mustParseFloat(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0
	}
	return float32(f)
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func mustParseDiv(s string) float32 {
	terms := strings.Split(s, "/")
	if len(terms) != 2 {
		return 0
	}
	n, err := strconv.ParseFloat(terms[0], 32)
	if err != nil {
		return 0
	}
	d, err := strconv.ParseFloat(terms[1], 32)
	if err != nil {
		return 0
	}
	return float32(n / d)
}

func GetVideoDuration(ctx context.Context, i string) (int64, error) {
	d, err := FFprobeFormatCommand(ctx, Duration, i)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(d, 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(v)), nil
}

func GetVideoSize(ctx context.Context, i string) (int64, error) {
	s, err := FFprobeFormatCommand(ctx, Size, i)
	if err != nil {
		return 0, err
	}
	sv, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return sv, nil
}

func FFprobeFormatCommand(ctx context.Context, format VideoFormate, i string) (string, error) {
	tCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	c := fmt.Sprintf("ffprobe -v error -show_entries format=%s -of default=noprint_wrappers=1:nokey=1", format)
	args := strings.Split(c, " ")

	o, err := exec.CommandContext(
		tCtx,
		args[0],
		append(args[1:], i)...,
	).CombinedOutput()

	if err != nil {
		return "", NewFFmpegGenError(err, string(o))
	}

	ov := strings.Trim(string(o), "\n")
	if ov == "" {
		return "", fmt.Errorf("failed to get %s", format)
	}

	return ov, nil
}

// Same as du -s -k go/ | grep -o '[[:digit:]]*'
func GetDirSize(ctx context.Context, dir string) (int64, error) {
	os := runtime.GOOS

	if os == "windows" {
		return 0, fmt.Errorf("unsupported os %s", os)
	}

	c := fmt.Sprintf("du -sk %s", dir)
	args := strings.Split(c, " ")

	cmd := exec.CommandContext(
		ctx,
		args[0],
		args[1:]...,
	)

	o, err := cmd.CombinedOutput()
	if err != nil {
		return 0, NewFFmpegGenError(err, string(o))
	}

	ov := strings.Trim(strings.Split(string(o), "\t")[0], "\n")
	if ov == "" {
		return 0, fmt.Errorf("failed to parse dir size")
	}

	sv, err := strconv.ParseInt(ov, 10, 64)
	if err != nil {
		return 0, err
	}

	bytes := sv * 1024

	return bytes, nil
}

type VideoResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (v *VideoResolution) String() string { return fmt.Sprintf("%dx%d", v.Width, v.Height) }

func GetResolution(ctx context.Context, input string) (*VideoResolution, error) {
	c := fmt.Sprintf("ffprobe -v error -select_streams v:0 -show_entries stream=width,height -of csv=s=x:p=0 %s", input)
	args := strings.Split(c, " ")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(o))
	}
	w, h, pE := parseRes(strings.Trim(string(o), "\n"))
	if pE != nil {
		return nil, pE
	}
	return &VideoResolution{w, h}, nil
}

func parseRes(res string) (int, int, error) {
	v := strings.Split(res, "x")
	if len(v) != 2 {
		return 0, 0, fmt.Errorf("invalid resolution %s", res)
	}
	w, err := strconv.Atoi(v[0])
	if err != nil {
		return 0, 0, err
	}
	h, err := strconv.Atoi(v[1])
	if err != nil {
		return 0, 0, err
	}
	return w, h, nil
}

type ThumbnailInput struct {
	// in seconds
	Duration int64
	Input    string
	Target   string
}

func GenThumbnail(ctx context.Context, in *ThumbnailInput) (string, error) {
	input := in.Input
	tf := in.Target
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	c := fmt.Sprintf(
		`ffmpeg -v error -y -ss %s -i %s -vframes 1 %s/thumbnail.jpg`,
		formatTimeSeconds(durTms(in.Duration)), input, tf,
	)
	args := strings.Split(c, " ")
	cmd := exec.CommandContext(tCtx, args[0], args[1:]...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return "", NewFFmpegGenError(err, string(o))
	}
	return fmt.Sprintf("%s/thumbnail.jpg", tf), nil
}

func durTms(duration int64) int { return int(math.RoundToEven(float64(duration) * 0.2)) }

// formatTimeSeconds formats seconds to hh:mm:ss
func formatTimeSeconds(sec int) string {
	h := sec / 3600
	m := (sec - h*3600) / 60
	s := sec - h*3600 - m*60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func GetAvailableVideoEncoders(ctx context.Context) ([]string, error) {
	// -hide_banner keeps output cleaner and stable for parsing
	out, err := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-encoders").CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(out))
	}
	encoders := make([]string, 0, 64)
	seen := make(map[string]struct{})
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Encoder rows begin with a 6-char flag block, e.g. "V....."
		// We only keep video encoders (first flag == 'V').
		fields := strings.Fields(line)
		if len(fields) < 2 || len(fields[0]) != 6 || fields[0][0] != 'V' {
			continue
		}
		name := fields[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		encoders = append(encoders, name)
	}
	slices.Sort(encoders)
	return encoders, nil
}
