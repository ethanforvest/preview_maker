# Preview Maker

CLI to generate previews for your videos with ease.

## Features

- Generate video previews at specified timestamps
- Customize the duration of each preview clip
- Automatically merge preview clips into a single output video
- Apply fade in/out effects to video and audio

## Requirements

- [ffmpeg](https://ffmpeg.org/)
- [ffprobe](https://ffmpeg.org/ffprobe.html)
- Go 1.15+

## Installation

1. Clone the repository:

```sh
git clone https://github.com/ethanforvest/preview_maker.git
cd preview_maker
```

2. Build the project:

```sh
go build . -o preview
```

### Alternatively, download the binaries from the [Releases](https://github.com/ethanforvest/preview_maker/releases) page.

## Usage

```sh
./preview <video-file> -t <timestamps> -d <duration>
```

### Arguments

- `video-file`: The path to the video file you want to process.
- `-t`: Timestamps (in percentages) to generate previews. Example: `-t 10 50 90` for 10%, 50%, and 90%.
- `-d`: Duration of each preview clip in seconds. Default is 10 seconds.

### Example

```sh
./preview myvideo.mp4 -t 10 50 90 -d 5
```

This command generates three preview clips of 5 seconds each at 10%, 50%, and 90% of the video duration and merges them into a single output file.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
