package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"os/exec"
	"strconv"
)

func getVideoDuration(videoPath string) (int, error) {
	// ffprobe command
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", videoPath)

	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	if err := cmd.Run(); err != nil {
		return -1, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	// Parse JSON output
	var result FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return -1, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	durFloat, err := strconv.ParseFloat(result.Format.Duration, 64)

	if err != nil {
		return 0, err
	}

	return int(durFloat), nil
}

func addMetaDataToDB(contentType string, path string, httpType string, video bool, thumbnailPath string) error {
	if video {

		if thumbnailPath == "" {
			return fmt.Errorf("thumbnailpath is empty for video")
		}

		file, err := os.Open(os.Getenv("MEDIAPATH") + thumbnailPath)

		if err != nil {
			return fmt.Errorf("could not find thumbnail given the path")
		}

		defer file.Close()

		img, _, err := image.DecodeConfig(file)
		if err != nil {
			fmt.Printf("Failed to decode image %s: %v\n", path, err)
			return err
		}

		query := "INSERT INTO media (type, path, size, mediatype, thumbnailPath, width, height) VALUES (?,?,?,?,?,?,?)"

		size, err := getFileSize(os.Getenv("MEDIAPATH") + path)

		if err != nil {
			fmt.Printf("Failed to get size")
			return err
		}

		result, err := db.Exec(query, contentType, path, size, httpType, thumbnailPath, img.Width, img.Height)

		if err != nil {
			return err
		}

		duration, err := getVideoDuration(os.Getenv("MEDIAPATH") + path)
		if err != nil {
			return err
		}

		vidId, err := result.LastInsertId()

		if err != nil {
			return err
		}

		insertDurationString := `INSERT INTO videoDuration(videoId, duration) VALUES (?,?)`

		_, insertErr := db.Exec(insertDurationString, vidId, duration)

		if insertErr != nil {
			return err
		}

		return nil

	}

	query := "INSERT INTO media (type, path, size, mediatype, width, height) VALUES (?,?,?,?,?,?)"

	size, sizeerr := getFileSize(os.Getenv("MEDIAPATH") + path)

	file, err := os.Open(os.Getenv("MEDIAPATH") + path)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", path, err)
		return nil
	}
	defer file.Close()

	// Decode the image to get its dimensions
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Printf("Failed to decode image %s: %v\n", path, err)
		return nil
	}

	if sizeerr != nil {
		return sizeerr
	}

	if _, err := db.Exec(query, contentType, path, size, httpType, img.Width, img.Height); err != nil {
		return err
	}

	return nil

}

func generateThumbnail(videoPath, fileName string) (string, error) {

	thumbnailPath := fileName + ".jpg"
	fullThumbnailPath := os.Getenv("THUMBNAIL") + thumbnailPath
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-frames:v", "1", "-q:v", "2", fullThumbnailPath)
	err := cmd.Run()
	return "thumbnails/" + thumbnailPath, err
}

func getFileSize(filePath string) (int64, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return -1, err
	}

	defer file.Close()
	fileStat, err := file.Stat()

	if err != nil {
		return -1, err
	}

	return fileStat.Size(), nil

}
