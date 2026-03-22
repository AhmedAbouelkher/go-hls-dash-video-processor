# go-hls-dash-video-processor

> **Notice:** This project is still under active development. APIs, CLI flags, and output behavior may change between versions.

A zero-dependency CLI tool written in Go that transcodes a source video into adaptive bitrate streaming formats — **HLS** and **DASH**. It uses **FFmpeg** for video/audio transcoding and **GPAC** for segment packaging, producing a ready-to-serve streaming output directory with multiple resolution variants, separated audio, and a thumbnail.

## [Demo](https://github.com/AhmedAbouelkher/go-hls-dash-video-processor/raw/refs/heads/main/demo.mp4)

<video controls>
  <source src="https://github.com/AhmedAbouelkher/go-hls-dash-video-processor/raw/refs/heads/main/demo.mp4" type="video/mp4">
</video>

## Features

- **Adaptive bitrate transcoding** — generates up to 5 resolution variants (`240p`, `360p`, `480p`, `720p`, `1080p`) with tuned CRF, bitrate, and buffer settings per tier
- **Smart resolution filtering** — automatically skips variants that exceed the source video's native resolution, avoiding unnecessary upscaling
- **Separate audio track** — extracts and encodes audio as AAC; handles both standard and MKV container inputs
- **Source integrity check** — validates the input file with FFmpeg before processing begins
- **Flexible packaging** — outputs **HLS** (`playlist.m3u8`) or combined **HLS + DASH** (`playlist.mpd`) via GPAC
- **MP4 fallback** — retains a muxed MP4 per resolution alongside the stream segments for direct playback compatibility
- **Thumbnail generation** — extracts a representative frame as a thumbnail image
- **Concurrent processing** — video transcoding and audio extraction run in parallel via a worker pool, reducing total processing time
- **Zero Go dependencies** — no external Go packages; relies only on the standard library

## Dependencies

The following tools must be installed and available in your `PATH`:

| Tool                         | Purpose                    | Install                                                                         |
| ---------------------------- | -------------------------- | ------------------------------------------------------------------------------- |
| [FFmpeg](https://ffmpeg.org) | Video/audio transcoding    | `brew install ffmpeg` / [ffmpeg.org/download](https://ffmpeg.org/download.html) |
| [GPAC](https://gpac.io)      | HLS/DASH segment packaging | `brew install gpac` / [gpac.io](https://gpac.io/downloads/)                     |

## Installation

```bash
git clone https://github.com/AhmedAbouelkher/go-hls-dash-video-processor.git
cd go-hls-dash-video-processor
```

### Build for your platform

**macOS**

```bash
GOOS=darwin GOARCH=amd64 go build -o video-processor .
```

**macOS (Apple Silicon)**

```bash
GOOS=darwin GOARCH=arm64 go build -o video-processor .
```

**Linux**

```bash
GOOS=linux GOARCH=amd64 go build -o video-processor .
```

**Windows**

```bash
GOOS=windows GOARCH=amd64 go build -o video-processor.exe .
```

Or run directly without building:

```bash
go run . <source_video> [options]
```

## Usage

```bash
video-processor <source_video> [options]
```

### Arguments

| Argument         | Description                                 |
| ---------------- | ------------------------------------------- |
| `<source_video>` | **(Required)** Path to the input video file |

### Options

| Flag        | Default                          | Description                                                                                               |
| ----------- | -------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `-d <path>` | Derived from the source filename | Path to the output directory where processed files will be written                                        |
| `-t <type>` | `hls`                            | Output packaging type. Accepted values: `hls`, `hls_and_dash`                                            |
| `-r <list>` | All resolutions                  | Comma-separated list of resolutions to generate. Accepted values: `240p`, `360p`, `480p`, `720p`, `1080p` |
| `-f`        | `false`                          | Force overwrite — removes the destination directory if it already exists                                  |

### Examples

Process a video with defaults (all resolutions, HLS output):

```bash
./video-processor my_video.mp4
```

Specify a custom output directory:

```bash
./video-processor my_video.mp4 -d /output/stream
```

Generate only specific resolutions:

```bash
./video-processor my_video.mp4 -r 480p,720p,1080p
```

Generate both HLS and DASH:

```bash
./video-processor my_video.mp4 -t hls_and_dash
```

Force overwrite an existing output directory:

```bash
./video-processor my_video.mp4 -f -d /output/stream
```

## Playing the Output

Once processing is complete, you can play the stream using any HLS-compatible player.

**mpv**

```bash
mpv ./bigbuckbunny/playlist.m3u8
```

**VLC or any HLS-compatible video player**

Open the player, then use "Open Network Stream" (or equivalent) and point it to the `playlist.m3u8` file path or URL if served over HTTP.

## Test Videos

Need a sample video to try the tool? A collection of public test videos of various sizes is available here:

[Public Test Videos](https://gist.github.com/AhmedAbouelkheir/22b357473a869c60aa4776f97a6767c6)

**Quick download example:**

```bash
curl -o sample.mp4 "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
./video-processor sample.mp4
```

## Roadmap

- [ ] Add hardware acceleration support
- [ ] Improve error handling
- [ ] Add unit tests

## License

This project is licensed under the [MIT License](LICENSE).  
Copyright © 2026 Ahmed Abouelkheir
