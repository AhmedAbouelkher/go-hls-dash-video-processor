package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ProcessType string

var (
	HLSProcessType  ProcessType = "hls"
	BOTHProcessType ProcessType = "hls_and_dash"
)

func isValidProcessType(s string) bool {
	switch s {
	case string(HLSProcessType):
		return true
	case string(BOTHProcessType):
		return true
	}
	return false
}

type videoComposeInput struct {
	videoID string
	vr      *videoResult
	ar      *audioResult
	typ     ProcessType
	target  string
}

func packageVideo(ctx context.Context, in *videoComposeInput) (*PackagingOutput, error) {
	tCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	ap := ""
	if in.ar != nil && in.ar.af != nil {
		ap = in.ar.af.Path
	}

	logger.WithFields(Fields{
		"video_id":          in.videoID,
		"target":            in.target,
		"resolutions_count": len(in.vr.resolutions),
		"resolution":        in.vr.vRes.String(),
		"audio":             ap,
		"type":              in.typ,
	}).Debug("📦 packaging video")

	vr := in.vr
	ar := in.ar
	dir := in.target

	aPath := ""
	if ar != nil && ar.af != nil {
		aPath = ar.af.Path
	}

	switch in.typ {
	case HLSProcessType:
		in := &GenerateHlsInput{
			Resolutions: vr.resolutions,
			Audio:       &aPath,
			Target:      dir,
		}
		return PackageHLS(tCtx, in)
	case BOTHProcessType:
		in := &GenerateHlsInput{
			Resolutions: vr.resolutions,
			Audio:       &aPath,
			Target:      dir,
		}
		// return packaging.PackageHLSAndDASH_v2(tCtx, in)
		return PackageHLSAndDASH(tCtx, in)
	default:
	}

	return nil, fmt.Errorf("unknown packaging type %s", in.typ)
}

type GenerateHlsInput struct {
	Resolutions   GenerationResolutions
	Audio         *string
	Target        string
	VideoDuration time.Duration
}

func (g *GenerateHlsInput) hasAudio() bool { return g.audioPath() != "" }

func (g *GenerateHlsInput) audioPath() string {
	if g.Audio == nil {
		return ""
	}
	return *g.Audio
}

type PackagingOutput struct {
	PackagingDuration time.Duration `json:"packaging_duration"`
}

func PackageHLS(ctx context.Context, in *GenerateHlsInput) (*PackagingOutput, error) {
	out := fmt.Sprintf("%s/playlist.m3u8", in.Target)
	inputs := []string{}
	if in.hasAudio() {
		inputs = append(
			inputs,
			fmt.Sprintf(
				"%s:#Bitrate=96kbps:#Template=audio/audio$Number$",
				in.audioPath(),
			),
		)
	}
	rInputs := in.Resolutions
	for _, r := range rInputs {
		v := r.VNoAudio
		res := r.Res
		i := fmt.Sprintf(
			"%s:#Template=%s/video$Number$",
			v,
			res.StringV2(),
		)
		inputs = append(inputs, i)
	}
	i := strings.Join(inputs, " -i ")
	c := fmt.Sprintf(
		"gpac -i %s -o %s:segdur=6:profile=dashavc264.live:muxtype=ts",
		i,
		out,
	)
	args := strings.Split(c, " ")
	t := time.Now()
	o, err := exec.CommandContext(ctx, args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(o))
	}

	return &PackagingOutput{
		PackagingDuration: time.Since(t),
	}, err
}

func PackageHLSAndDASH(ctx context.Context, in *GenerateHlsInput) (*PackagingOutput, error) {
	out := fmt.Sprintf("%s/playlist.mpd", in.Target)
	inputs := []string{}
	if in.hasAudio() {
		inputs = append(
			inputs,
			fmt.Sprintf(
				"%s:#Bitrate=96kbps:#Template=audio/audio$Number$",
				in.audioPath(),
			),
		)
	}
	rInputs := in.Resolutions
	for _, r := range rInputs {
		v := r.VNoAudio
		res := r.Res
		i := fmt.Sprintf(
			"%s:#Template=%s/video$Number$",
			v,
			res.StringV2(),
		)
		inputs = append(inputs, i)
	}
	i := strings.Join(inputs, " -i ")
	c := fmt.Sprintf(
		"gpac -i %s -o %s:segdur=10:dual:profile=dashavc264.live:muxtype=ts",
		i,
		out,
	)
	args := strings.Split(c, " ")
	t := time.Now()
	o, err := exec.CommandContext(ctx, args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(o))
	}
	return &PackagingOutput{
		PackagingDuration: time.Since(t),
	}, nil
}

func PackageHLSAndDASH_v2(ctx context.Context, in *GenerateHlsInput) (*PackagingOutput, error) {
	c, err := constructCMD(in)
	if err != nil {
		return nil, err
	}
	args := strings.Split(strings.TrimPrefix(c, " "), " ")
	t := time.Now()
	o, err := exec.CommandContext(ctx, "packager", args...).CombinedOutput()
	if err != nil {
		return nil, NewFFmpegGenError(err, string(o))
	}
	return &PackagingOutput{
		PackagingDuration: time.Since(t),
	}, err
}

func constructCMD(in *GenerateHlsInput) (string, error) {
	targetPath := in.Target
	resolutions := in.Resolutions
	var cmd strings.Builder
	if len(resolutions) == 0 {
		return "", fmt.Errorf("no resolutions to package")
	}
	if in.hasAudio() {
		latestRes := resolutions[len(resolutions)-1]
		cmd.WriteString(fmt.Sprintf(
			" in=%s,stream=audio,segment_template=%s/audio/$Number$.aac,playlist_name=%s/audio/playlist.m3u8,hls_group_id=audio,hls_name=ar",
			latestRes.Video,
			targetPath,
			targetPath,
		))
	}
	for _, res := range resolutions {
		noAudio := res.VNoAudio
		name := res.Res.Name()
		tt := fmt.Sprintf("%s/video_%s", targetPath, name)
		cmd.WriteString(fmt.Sprintf(
			" in=%s,stream=video,segment_template=%s/$Number$.ts,playlist_name=%s/playlist.m3u8,iframe_playlist_name=%s/iframe.m3u8",
			noAudio,
			tt,
			tt,
			tt,
		))
	}
	cmd.WriteString(fmt.Sprintf(" --mpd_output %s/playlist.mpd", targetPath))
	cmd.WriteString(fmt.Sprintf(" --hls_master_playlist_output %s/playlist.m3u8", targetPath))
	return cmd.String(), nil

}
