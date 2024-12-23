package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
)

// Helper function to execute ffmpeg commands
func runCommand(cmd []string) error {
	fmt.Println("Running command:", cmd)
	err := exec.Command(cmd[0], cmd[1:]...).Run()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// Get video duration using ffprobe
func getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

var args struct {
	File       string  `arg:"positional,required" help:"The File To Process"`
	TimeStamps []uint8 `arg:"-t" help:"The TimeStamps To Process ex: -t 10 20 70 (10% 20% 70%) of the file"`
	Duration   int     `default:"10" arg:"-d" help:"The Duration of each clip in seconds"`
}

func init() {
	args.TimeStamps = []uint8{10, 50, 90}
}

func main() {
	arg.MustParse(&args)
	// fmt.Println(args.File, args.TimeStamps)
	video_path_input := args.File

	// Step 1: Get video file path
	videoPath := strings.TrimSpace(video_path_input)

	// Step 2: Get duration of each clip
	clipDuration := args.Duration

	// Step 3: Get video duration
	videoDuration, err := getVideoDuration(videoPath)
	if err != nil {
		fmt.Println("Error getting video duration:", err)
		return
	}

	// Step 4: Define/Parse clip start points (percentages)
	clipPoints := []float64{}
	for _, point := range args.TimeStamps {
		clipPoints = append(clipPoints, float64(point)/100)
	}

	tempFiles := []string{}

	for i, point := range clipPoints {
		startTime := videoDuration * point
		tempFile := fmt.Sprintf("clip_%d.mp4", i+1)
		tempFiles = append(tempFiles, tempFile)

		// Add transition only to the first clip's AUDIO and not video
		if i == 0 {
			// Adjust fade start time relative to each clip's start time
			afade := fmt.Sprintf("afade=t=in:st=%.2f:d=0.5", startTime) // Start audio fade relative to clip start

			cmd := []string{"ffmpeg", "-i", videoPath, "-af", afade, "-ss", fmt.Sprintf("%.2f", startTime), "-t", fmt.Sprintf("%d", clipDuration), "-c:v", "libx264", "-crf", "18", "-preset", "slow", "-c:a", "aac", "-b:a", "192k", "-y", tempFile}

			if err := runCommand(cmd); err != nil {
				fmt.Println("Error creating clip:", err)
				return
			}

			continue
		}

		// Adjust fade start time relative to each clip's start time
		vfade := fmt.Sprintf("fade=t=in:st=%.2f:d=0.5", startTime)  // Start fade relative to clip start
		afade := fmt.Sprintf("afade=t=in:st=%.2f:d=0.5", startTime) // Start audio fade relative to clip start

		cmd := []string{"ffmpeg", "-i", videoPath, "-vf", vfade, "-af", afade, "-ss", fmt.Sprintf("%.2f", startTime), "-t", fmt.Sprintf("%d", clipDuration), "-c:v", "libx264", "-crf", "18", "-preset", "slow", "-c:a", "aac", "-b:a", "192k", "-y", tempFile}

		if err := runCommand(cmd); err != nil {
			fmt.Println("Error creating clip:", err)
			return
		}
	}

	// Step 6: Merge clips into one file
	mergeFile := "file_list.txt"
	outputFile := "output.mp4"
	fileList, err := os.Create(mergeFile)
	if err != nil {
		fmt.Println("Error creating file list:", err)
		return
	}
	defer fileList.Close()

	for _, tempFile := range tempFiles {
		fileList.WriteString(fmt.Sprintf("file '%s'\n", tempFile))
	}

	cmd := []string{"ffmpeg", "-f", "concat", "-safe", "0", "-i", mergeFile, "-c", "copy", "-y", outputFile}
	if err := runCommand(cmd); err != nil {
		fmt.Println("Error merging clips:", err)
		goto clean
	}

	fmt.Println("Output video saved as:", outputFile)

	// Cleanup temporary files
clean:
	err = os.Remove(mergeFile)
	if err != nil {
		fmt.Println("Error removing file list: ", err)
	}

	for _, tempFile := range tempFiles {
		os.Remove(tempFile)
	}
}
