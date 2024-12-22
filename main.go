package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Helper function to execute ffmpeg commands
func runCommand(cmd string) error {
	fmt.Println("Running command:", cmd)
	parts := strings.Split(cmd, " ")
	err := exec.Command(parts[0], parts[1:]...).Run()
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

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("USAGE: %s <file_path>\n", os.Args[0])
		os.Exit(1)
	}

	video_path_input := os.Args[1]

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Get video file path
	videoPath := strings.TrimSpace(video_path_input)

	// Step 2: Get duration of each clip
	fmt.Print("Enter the duration (in seconds) for each clip: ")
	durationInput, _ := reader.ReadString('\n')
	clipDuration, err := strconv.Atoi(strings.TrimSpace(durationInput))
	if err != nil {
		fmt.Println("Invalid duration:", err)
		return
	}

	// Step 3: Get video duration
	videoDuration, err := getVideoDuration(videoPath)
	if err != nil {
		fmt.Println("Error getting video duration:", err)
		return
	}

	// Step 4: Define clip start points (percentages)
	clipPoints := []float64{0.1, 0.5, 0.9}
	tempFiles := []string{}

	for i, point := range clipPoints {
		startTime := videoDuration * point
		tempFile := fmt.Sprintf("clip_%d.mp4", i+1)
		tempFiles = append(tempFiles, tempFile)

		// Add transition only to the AUDIO and not video
		if i == 0 {
			// Adjust fade start time relative to each clip's start time
			afade := fmt.Sprintf("afade=t=in:st=%.2f:d=0.5", startTime) // Start audio fade relative to clip start

			cmd := fmt.Sprintf("ffmpeg -i %s -af %s -ss %.2f -t %d -c:v libx264 -crf 18 -preset slow -c:a aac -b:a 192k -y %s",
				videoPath, afade, startTime, clipDuration, tempFile)

			if err := runCommand(cmd); err != nil {
				fmt.Println("Error creating clip:", err)
				return
			}

			continue
		}

		// Adjust fade start time relative to each clip's start time
		vfade := fmt.Sprintf("fade=t=in:st=%.2f:d=0.5", startTime)  // Start fade relative to clip start
		afade := fmt.Sprintf("afade=t=in:st=%.2f:d=0.5", startTime) // Start audio fade relative to clip start

		cmd := fmt.Sprintf("ffmpeg -i %s -vf %s -af %s -ss %.2f -t %d -c:v libx264 -crf 18 -preset slow -c:a aac -b:a 192k -y %s",
			videoPath, vfade, afade, startTime, clipDuration, tempFile)

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

	cmd := fmt.Sprintf("ffmpeg -f concat -safe 0 -i %s -c copy -y %s", mergeFile, outputFile)
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
