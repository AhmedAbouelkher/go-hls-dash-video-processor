package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	Res240p  = "240p"
	Res360p  = "360p"
	Res480p  = "480p"
	Res720p  = "720p"
	Res1080p = "1080p"
)

var (
	_resolutions = []resolution{
		{ID: Res240p, Width: 426, Height: 240, crf: 23, bitrate: 400, maxBitrate: 400, bufsize: 800, timeout: 30 * time.Minute},               // 240p
		{ID: Res360p, Width: 640, Height: 360, crf: 25, bitrate: 800, maxBitrate: 800, bufsize: 1500, timeout: time.Hour},                     // 360p
		{ID: Res480p, Width: 854, Height: 480, crf: 28, bitrate: 1000, maxBitrate: 1000, bufsize: 2000, timeout: time.Hour + 10*time.Minute},  // 480p
		{ID: Res720p, Width: 1280, Height: 720, crf: 28, bitrate: 1500, maxBitrate: 2500, bufsize: 3000, timeout: time.Hour + 40*time.Minute}, // 720p
		{ID: Res1080p, Width: 1920, Height: 1080, crf: 28, bitrate: 3500, maxBitrate: 5000, bufsize: 5000, timeout: 2 * time.Hour},            // 1080p
	}
)

func isValidResolution(r string) bool {
	for _, res := range _resolutions {
		if res.ID == r {
			return true
		}
	}
	return false
}

// MARK:- Video Generation

type videoResult struct {
	resolutions GenerationResolutions
	vRes        *VideoResolution
	err         error
}

type videoGenInput struct {
	videoID  string
	settings *ResolutionsSettings
	input    string
	target   string
	size     int64
	Duration int64
}

func generateVideo(ctx context.Context, in *videoGenInput) (GenerationResolutions, *VideoResolution, error) {
	tCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	res, err := GetResolution(tCtx, in.input)
	if err != nil {
		return nil, nil, err
	}

	logger.WithFields(Fields{
		"video_id":         in.videoID,
		"video_resolution": res.String(),
		"input":            in.input,
		"target":           in.target,
	}).Debugf("🔧 source width x height: %dx%d", res.Width, res.Height)

	resIn := &GenerateResolutionsInput{
		VideoID:  in.videoID,
		Res:      res,
		Settings: in.settings,
		Input:    in.input,
		Target:   in.target,
		Size:     in.size,
		Duration: in.Duration,
	}
	resolutions, err := GenerateResolutions(ctx, resIn)
	if err != nil {
		return nil, nil, err
	}
	return resolutions, res, nil
}

// MARK:- Multiple Resolutions Generation

type ResolutionsSettings struct {
	DisableMP4Fallback bool     `json:"disable_mp4_fallback"`
	EnabledResolutions []string `json:"enabled_resolutions"`
	Bitrate240p        int      `json:"bitrate_240p"`
	Bitrate360p        int      `json:"bitrate_360p"`
	Bitrate480p        int      `json:"bitrate_480p"`
	Bitrate720p        int      `json:"bitrate_720p"`
	Bitrate1080p       int      `json:"bitrate_1080p"`
}

func (p *ResolutionsSettings) LogDetails() Fields {
	return Fields{
		"disable_mp4_fallback": p.DisableMP4Fallback,
		"enabled_resolutions":  strings.Join(p.EnabledResolutions, ","),
		"bitrate_240p":         p.Bitrate240p,
		"bitrate_360p":         p.Bitrate360p,
		"bitrate_480p":         p.Bitrate480p,
		"bitrate_720p":         p.Bitrate720p,
		"bitrate_1080p":        p.Bitrate1080p,
	}
}

type GenerateResolutionsInput struct {
	VideoID  string
	Res      *VideoResolution
	Settings *ResolutionsSettings
	Input    string
	Target   string
	Size     int64
	Duration int64
}

func GenerateResolutions(ctx context.Context, in *GenerateResolutionsInput) (GenerationResolutions, error) {
	settings := in.Settings

	if settings == nil {
		return nil, fmt.Errorf("settings is required")
	}

	resolutions := []resolution{}
	resMap := resolutionsMap()
	for _, r := range settings.EnabledResolutions {
		srcRes := in.Res
		// skip if resolution not found
		// skip if resolution is bigger than source resolution
		res, ok := resMap[r]
		if !ok || res.Height > srcRes.Height {
			continue
		}
		maxBtr := res.maxBitrate
		switch res.ID {
		case Res240p:
			maxBtr = lgZero(settings.Bitrate240p, maxBtr)
		case Res360p:
			maxBtr = lgZero(settings.Bitrate360p, maxBtr)
		case Res480p:
			maxBtr = lgZero(settings.Bitrate480p, maxBtr)
		case Res720p:
			maxBtr = lgZero(settings.Bitrate720p, maxBtr)
		case Res1080p:
			maxBtr = lgZero(settings.Bitrate1080p, maxBtr)
		}
		res.maxBitrate = maxBtr
		resolutions = append(resolutions, res)
	}

	resToGenerate := []string{}
	for _, r := range resolutions {
		resToGenerate = append(resToGenerate, r.StringV2())
	}
	logger.WithFields(Fields{
		"video_id":         in.VideoID,
		"video_resolution": in.Res.String(),
		"input":            in.Input,
		"target":           in.Target,
		"size":             in.Size,
		"duration":         formatTimeSeconds(int(in.Duration)),
		"resolutions":      strings.Join(resToGenerate, ", "),
	}).Debugf("⚙️ generating %d resolutions", len(resolutions))

	if len(resolutions) == 0 {
		return nil, fmt.Errorf("failed to find suitable processing resolutions with %s", in.Res.String())
	}

	res := GenerationResolutions{}

	for _, r := range resolutions {
		func() {
			timeout := r.CalcTimeout(&TimeoutCalc{size: in.Size, duration: in.Duration})

			rCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			now := time.Now()
			out, err := generateRes(rCtx, &generateResInput{
				r:      &r,
				input:  in.Input,
				target: in.Target,
			})

			if err != nil {
				logger.WithError(err).WithFields(Fields{
					"video_id":      in.VideoID,
					"res":           in.Res.String(),
					"timeout_sec":   timeout / time.Second,
					"time_took_sec": time.Since(now).Seconds(),
					"input":         in.Input,
					"target":        in.Target,
					"resolution":    in.Res,
					"failed":        r.String(),
				}).Error("🥲 failed to generate resolution")
				return
			}

			res = append(res, out)
		}()
	}

	return res, nil
}

// MARK:- Single Resolution Generation

type generateResInput struct {
	r      *resolution
	input  string
	target string
}

type GenerationResolutions []*GenerateResOutput

func (g GenerationResolutions) Resolutions() []string {
	var out []string
	for _, r := range g {
		out = append(out, r.Res.Name())
	}
	return out
}

func (g GenerationResolutions) LogDetails() Fields {
	o := Fields{}
	for _, r := range g {
		o[r.Res.StringV2()] = fmt.Sprintf("video=%s generation_duration_sec=%.2f audio_removal_duration_sec=%.2f", r.Video, r.GenerationDuration.Seconds(), r.AudioRemovalDuration.Seconds())
	}
	return o
}

type GenerateResOutput struct {
	Res      *resolution `json:"res"`
	Video    string      `json:"video"`
	VNoAudio string      `json:"v_no_audio"`

	GenerationDuration   time.Duration `json:"generation_duration"`
	AudioRemovalDuration time.Duration `json:"audio_removal_duration"`
}

func generateRes(ctx context.Context, in *generateResInput) (*GenerateResOutput, error) {
	t1 := time.Now()
	rIn := &generateInput{
		r:      in.r,
		input:  in.input,
		target: in.target,
	}
	rOut, err := generate(ctx, rIn)
	if err != nil {
		return nil, err
	}
	genT := time.Since(t1)

	t2 := time.Now()
	rAIn := &removeAudioInput{
		r:      in.r,
		input:  rOut,
		target: in.target,
	}
	rAOut, err := removeAudio(ctx, rAIn)
	if err != nil {
		return nil, err
	}
	removeAudioT := time.Since(t2)

	return &GenerateResOutput{
		Res:      in.r,
		Video:    rOut,
		VNoAudio: rAOut,

		GenerationDuration:   genT,
		AudioRemovalDuration: removeAudioT,
	}, nil
}

// MARK:- Video Resolutions

type resolution struct {
	ID     string
	Width  int
	Height int
	crf    int

	bitrate    int
	maxBitrate int
	bufsize    int

	timeout time.Duration
}

// String returns the resolution in the format of 1280:720
func (r *resolution) String() string { return fmt.Sprintf("%d:%d", r.Width, r.Height) }

// StringV2 returns the resolution in the format of 1280x720
func (r *resolution) StringV2() string { return fmt.Sprintf("%dx%d", r.Width, r.Height) }

// Name returns the resolution in the format of 720p
func (r *resolution) Name() string { return fmt.Sprintf("%dp", r.Height) }

func resolutionsMap() map[string]resolution {
	m := make(map[string]resolution)
	for _, r := range _resolutions {
		m[r.ID] = r
	}
	return m
}

type TimeoutCalc struct {
	size     int64
	duration int64
}

func (r *resolution) CalcTimeout(cal *TimeoutCalc) time.Duration {
	sto := r.sizeTimeout(cal.size)
	dto := r.durationTimeout(cal.duration)
	if sto > dto {
		return sto
	}
	return dto
}

func (r *resolution) sizeTimeout(size int64) time.Duration {
	gb := bytesToGB(size)
	if gb <= 1 {
		return r.timeout
	}
	// 1 gb += 10 minute
	t := int(math.Ceil(gb)) * 10
	return r.timeout + time.Duration(t)*time.Minute
}
func bytesToGB(b int64) float64 { return float64(b) * 1e-9 }

func (r *resolution) durationTimeout(duration int64) time.Duration {
	hr := secondsToHours(duration)
	if hr <= 1.5 {
		return r.timeout
	}
	// 1 hr += 10 minute
	t := hr * 10
	return r.timeout + time.Duration(t)*time.Minute
}

func secondsToHours(seconds int64) float64 { return float64(seconds) / 3600 }

func lgZero(v, dft int) int {
	if v == 0 {
		return dft
	}
	return v
}
