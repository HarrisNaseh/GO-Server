package main

import (
	// "net/http"

	"database/sql"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// imageTest()

	db = dbInit()

	defer db.Close()

	router := gin.Default()

	router.MaxMultipartMemory = 8 << 20

	router.Use(corsMiddleware())
	// go router.GET("/video", getVideo)
	go router.GET("/media/:id", getMediaById)
	go router.GET("/", getAll)
	go router.GET("/media/:id/thumbnail", getThumbnailById)

	//post request
	go router.POST("/upload", uploadFiles)

	//detele route
	go router.DELETE("/media/:id", deleteById)
	// router.Run() //Run this when you want this to run on the network
	router.Run(":8080")
}

func uploadFiles(c *gin.Context) {

	uploadFile := func(file *multipart.FileHeader, path string) error {
		if err := c.SaveUploadedFile(file, path); err != nil {
			return err
		}

		return nil
	}

	form, _ := c.MultipartForm()

	files := form.File["files"]
	filesUploaded := 0
	errorString := ""
	basePath := os.Getenv("MEDIAPATH")
	for _, file := range files {

		ext := strings.ToLower(filepath.Ext(file.Filename))

		if mimi, vaild := imageFormats[ext]; vaild {

			path := basePath + "images/" + file.Filename

			if err := uploadFile(file, path); err != nil {
				// c.String(http.StatusInternalServerError, "Failed to upload file to server")
				errorString = fmt.Sprintf("%sFailed to uploaded file%s\n", errorString, file.Filename)
				continue
			}

			if err := addMetaDataToDB("image", "images/"+file.Filename, mimi, false, ""); err != nil {
				// c.String(http.StatusInternalServerError, "Could not add metadata related to image to server")
				os.Remove(path)
				errorString = fmt.Sprintf("%sFailed to uploaded file%s\n", errorString, file.Filename)
				//if this fails we should delete the media
				continue
			}

			filesUploaded++

		} else if mimi, valid := videoFormats[ext]; valid {

			path := basePath + "videos/" + file.Filename
			if err := uploadFile(file, path); err != nil {
				// c.String(http.StatusInternalServerError, fmt.Sprintf("Could not upload %s to server", file.Filename))
				errorString = fmt.Sprintf("%sFailed to uploaded file%s\n", errorString, file.Filename)
				continue
			}

			name := strings.Split(file.Filename, ".")[0]
			thumbnailPath, err := generateThumbnail(path, name)

			if err != nil {
				// c.String(http.StatusInternalServerError, fmt.Sprintf("Could not generate thumbnail for %s", file.Filename))
				os.Remove(path)
				errorString = fmt.Sprintf("%sCould not generate thumbnail for%s\n", errorString, file.Filename)
				continue
			}

			if err := addMetaDataToDB("video", "videos/"+file.Filename, mimi, true, thumbnailPath); err != nil {
				// c.String(http.StatusInternalServerError, fmt.Sprintf("Could not add metadata for %s to server", file.Filename))
				os.Remove(path)
				os.Remove(basePath + thumbnailPath)
				errorString = fmt.Sprintf("%sFailed to uploaded file%s\n", errorString, file.Filename)
				continue
			}
			filesUploaded++

		} else {
			// c.String(http.StatusInternalServerError, fmt.Sprintf("Could not upload %s to server because file type is not supported", file.Filename))
			errorString = fmt.Sprintf("%sCould not upload %s to server because file type is not supported\n", errorString, file.Filename)
		}

	}

	//TODO: Return a Status error if errorString is not empty
	c.JSON(http.StatusOK, gin.H{
		"Files_Uploaded": filesUploaded,
		"Error_Strings":  errorString,
	})

}

func getThumbnailById(c *gin.Context) {

	id := c.Param("id")
	row := db.QueryRow("SELECT type, thumbnailPath, path, size FROM media WHERE id=?", id)

	var media Media

	if err := row.Scan(&media.TYPE, &media.thumbnailPath, &media.PATH, &media.Size); err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, fmt.Sprintf("No media exists with id: %s", id))
			return
		}
		c.String(http.StatusInternalServerError, fmt.Sprintf("mediaById %s: %v", id, err))
	}

	if media.TYPE == "video" {
		file, err := os.Open(os.Getenv("MEDIAPATH") + *media.thumbnailPath)

		if err != nil {
			c.String(http.StatusNotFound, "Image not found.")
			return
		}
		defer file.Close()

		fileStat, err := file.Stat()

		if err != nil {
			c.String(http.StatusInternalServerError, "Could Not Stat File.")
			return
		}

		c.Header("Content-Length", fmt.Sprintf("%d", fileStat.Size()))
		c.Header("Content-Type", "image/jpg")

		io.Copy(c.Writer, file)
		return
	}

	getImageById(c, media)

}

func getVideoById(c *gin.Context, media Media) {
	file, err := os.Open(os.Getenv("MEDIAPATH") + media.PATH)

	if err != nil {
		c.String(http.StatusNotFound, "Video not found.")
		return
	}

	defer file.Close()

	fileSize := media.Size

	rangeHeader := c.GetHeader("Range")

	if rangeHeader == "" {
		c.Header("Content-Type", media.MediaType)
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Length", fmt.Sprint(fileSize))
		io.Copy(c.Writer, file)
	}

	var start int64
	var end int64
	_, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

	if err != nil {
		_, err = fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
		if err != nil {
			var diff int64
			_, err = fmt.Sscanf(rangeHeader, "bytes=-%d", &diff)
			if err != nil {
				c.String(http.StatusInternalServerError, "Could not resolve range")
			}
			start = fileSize - 1 - diff
		}
		end = fileSize - 1

	}

	if end >= fileSize || start > end || start < 0 || end < 0 {

		c.Status(http.StatusRequestedRangeNotSatisfiable)

	} else {
		c.Status(http.StatusPartialContent)
		c.Header("Content-Type", media.MediaType)
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))

		file.Seek(start, io.SeekStart)
		io.CopyN(c.Writer, file, end-start+1)
	}

}

func getMediaById(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT id, type, path, size, mediatype, thumbnailPath FROM media WHERE id=?", id)

	var media Media

	if err := row.Scan(&media.ID, &media.TYPE, &media.PATH, &media.Size, &media.MediaType, &media.thumbnailPath); err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, fmt.Sprintf("No media exists with id: %s", id))
			return
		}
		c.String(http.StatusInternalServerError, fmt.Sprintf(" Problem with %s: %v", id, err))
	}

	if media.TYPE == "image" {
		getImageById(c, media)
		return
	}

	getVideoById(c, media)

}

func getImageById(c *gin.Context, media Media) {

	file, err := os.Open(os.Getenv("MEDIAPATH") + media.PATH)

	if err != nil {
		c.String(http.StatusNotFound, "Image not found.")
		return
	}

	defer file.Close()

	c.Header("Content-Length", fmt.Sprintf("%d", media.Size))
	c.Header("Content-Type", media.MediaType)

	io.Copy(c.Writer, file)
}

func getAll(c *gin.Context) {
	rows, err := db.Query("SELECT id, type, width, height, duration FROM media left JOIN videoduration ON media.id=videoduration.videoid")

	if err != nil {
		c.String(http.StatusInternalServerError, "Could not get all rows from database")
		return
	}

	defer rows.Close()

	var items []gin.H

	for rows.Next() {
		var media Media

		if err := rows.Scan(&media.ID, &media.TYPE, &media.WIDTH, &media.HEIGHT, &media.DURATION); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
		}

		items = append(items, gin.H{
			"id":       media.ID,
			"type":     media.TYPE,
			"width":    media.WIDTH,
			"height":   media.HEIGHT,
			"duration": media.DURATION,
		})
	}

	if err := rows.Err(); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
	}

	c.JSON(http.StatusOK, items)
}

func deleteById(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT path, thumbnailPath, type FROM media WHERE id=?", id)

	var media Media

	if err := row.Scan(&media.PATH, &media.thumbnailPath, &media.TYPE); err != nil {
		c.String(http.StatusNotFound, "Could not find file to delete")
		return
	}

	basePath := os.Getenv("MEDIAPATH")
	if err := os.Remove(basePath + media.PATH); err != nil {
		c.String(http.StatusNotFound, "Path not found")
		return
	}

	if media.TYPE == "video" {
		if err := os.Remove(basePath + *media.thumbnailPath); err != nil {
			c.String(http.StatusNotFound, "Thumbnail Path not found")
			return
		}
	}

	db.Exec("DELETE FROM media WHERE id=?", id)

}
