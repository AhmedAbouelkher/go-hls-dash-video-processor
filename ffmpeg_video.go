package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type generateInput struct {
	r      *resolution
	input  string
	target string
}

// FROM: https://trac.ffmpeg.org/wiki/Encode/H.264
func generate(ctx context.Context, in *generateInput) (string, error) {
	outF := fmt.Sprintf("%s/play_%s.mp4", in.target, in.r.Name())

	c := fmt.Sprintf(
		`ffmpeg -v error -y -i %s -c:v libx264 -preset medium -crf %d -b:v %dk -maxrate %dk -bufsize %dk -c:a copy -vf scale=-2:%d -f mp4 %s`,
		in.input,
		in.r.crf,
		in.r.bitrate,
		in.r.maxBitrate,
		in.r.bufsize,
		in.r.Height,
		outF,
	)

	args := strings.Split(c, " ")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return "", NewFFmpegGenError(err, string(o))
	}
	return outF, nil
}
