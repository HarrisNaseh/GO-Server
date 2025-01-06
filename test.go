package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

func imageTest() {

	dirPath := os.Getenv("MEDIAPATH") + "/images"
	// Iterate over files in the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		file, err := os.Open(path)
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

		// Print the image dimensions
		fmt.Printf("File: %s, Width: %d, Height: %d\n", path, img.Width, img.Height)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directory: %v\n", err)
	}

	// file, err := os.Open(os.Getenv("MEDIAPATH") + "/images/1734053180824.jpg")

	// if err != nil {
	// 	// c.String(http.StatusNotFound, "Image not found.")
	// 	return
	// }

	// defer file.Close()

}
