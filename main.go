package main

import (
	// "net/http"

	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var imageFormats = map[string]string{
	".bmp":  "image/bmp",
	".btif": "image/prs.btif",
	".cgm":  "image/cgm",
	".cmx":  "image/x-cmx",
	".djv":  "image/vnd.djvu",
	".djvu": "image/vnd.djvu",
	".dwg":  "image/vnd.dwg",
	".dxf":  "image/vnd.dxf",
	".fbs":  "image/vnd.fastbidsheet",
	".fh":   "image/x-freehand",
	".fh4":  "image/x-freehand",
	".fh5":  "image/x-freehand",
	".fh7":  "image/x-freehand",
	".fhc":  "image/x-freehand",
	".fpx":  "image/vnd.fpx",
	".fst":  "image/vnd.fst",
	".g3":   "image/g3fax",
	".gif":  "image/gif",
	".ico":  "image/x-icon",
	".ief":  "image/ief",
	".jpe":  "image/jpeg",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".mdi":  "image/vnd.ms-modi",
	".mmr":  "image/vnd.fujixerox.edmics-mmr",
	".npx":  "image/vnd.net-fpx",
	".pbm":  "image/x-portable-bitmap",
	".pct":  "image/x-pict",
	".pcx":  "image/x-pcx",
	".pgm":  "image/x-portable-graymap",
	".pic":  "image/x-pict",
	".png":  "image/png",
	".pnm":  "image/x-portable-anymap",
	".ppm":  "image/x-portable-pixmap",
	".psd":  "image/vnd.adobe.photoshop",
	".ras":  "image/x-cmu-raster",
	".rgb":  "image/x-rgb",
	".rlc":  "image/vnd.fujixerox.edmics-rlc",
	".svg":  "image/svg+xml",
	".svgz": "image/svg+xml",
	".tif":  "image/tiff",
	".tiff": "image/tiff",
	".wbmp": "image/vnd.wap.wbmp",
	".xbm":  "image/x-xbitmap",
	".xif":  "image/vnd.xiff",
	".xpm":  "image/x-xpixmap",
	".xwd":  "image/x-xwindowdump",
}

var videoFormats = map[string]string{
	".avi":   "x-msvideo",
	".ogv":   "video/ogg",
	".ts":    "mp2t",
	".3g2":   "video/3gpp2",
	".3gp":   "video/3gpp",
	".asf":   "video/x-ms-asf",
	".asx":   "video/x-ms-asf",
	".f4v":   "video/x-f4v",
	".fli":   "video/x-fli",
	".flv":   "video/x-flv",
	".fvt":   "video/vnd.fvt",
	".h261":  "video/h261",
	".h263":  "video/h263",
	".h264":  "video/h264",
	".jpgm":  "video/jpm",
	".jpgv":  "video/jpeg",
	".jpm":   "video/jpm",
	".m1v":   "video/mpeg",
	".m2v":   "video/mpeg",
	".m4u":   "video/vnd.mpegurl",
	".m4v":   "video/x-m4v",
	".mj2":   "video/mj2",
	".mjp2":  "video/mj2",
	".mov":   "video/quicktime",
	".movie": "video/x-sgi-movie",
	".mp4":   "video/mp4",
	".mp4v":  "video/mp4",
	".mpa":   "video/mpeg",
	".mpeg":  "video/mpeg",
	".mpg":   "video/mpeg",
	".mpg4":  "video/mp4",
	".mxu":   "video/vnd.mpegurl",
	".pyv":   "video/vnd.ms-playready.media.pyv",
	".qt":    "video/quicktime",
	".viv":   "video/vnd.vivo",
	".wm":    "video/x-ms-wm",
	".wmv":   "video/x-ms-wmv",
	".wmx":   "video/x-ms-wmx",
	".wvx":   "video/x-ms-wvx",
}

type Media struct {
	ID            int    `json:"id"`
	TYPE          string `json:"type"`
	PATH          string
	DATE          time.Time
	MediaType     string `json:"mediatype"`
	Size          int64  `json:"size"`
	thumbnailPath *string
	WIDTH         int  `json:"width"`
	HEIGHT        int  `json:"height"`
	DURATION      *int `json:"duration"`
}

type FFProbeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
	} `json:"format"`
}

var db *sql.DB

// Store this inside another table. Probably two tables and then combine
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

// CORS middleware function definition
func corsMiddleware() gin.HandlerFunc {
	// Define allowed origins as a comma-separated string
	originsString := "http://localhost:5173,http://10.0.0.211:5173"
	var allowedOrigins []string
	if originsString != "" {
		// Split the originsString into individual origins and store them in allowedOrigins slice
		allowedOrigins = strings.Split(originsString, ",")
	}

	// Return the actual middleware handler function
	return func(c *gin.Context) {
		// Function to check if a given origin is allowed
		isOriginAllowed := func(origin string, allowedOrigins []string) bool {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		}

		// Get the Origin header from the request
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		if isOriginAllowed(origin, allowedOrigins) {
			// If the origin is allowed, set CORS headers in the response
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		}

		// Handle preflight OPTIONS requests by aborting with status 204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Call the next handler
		c.Next()
	}
}

func getThumbnailById(c *gin.Context) {

	id := c.Param("id")
	row := db.QueryRow("SELECT type, thumbnailPath, path FROM media WHERE id=?", id)

	var media Media

	if err := row.Scan(&media.TYPE, &media.thumbnailPath, &media.PATH); err != nil {
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

func dbInit() *sql.DB {
	dbName := os.Getenv("DBURL")
	db, err := sql.Open("sqlite3", dbName)

	if err != nil {
		log.Fatalf("Can not connect to database %s", err)
	}

	createString := `CREATE TABLE IF NOT EXISTS media(
        id INTEGER PRIMARY KEY NOT NULL,
        type TEXT NOT NULL,
        path TEXT NOT NULL UNIQUE,
        timestamp DATE DEFAULT CURRENT_TIMESTAMP,
        size INTEGER NOT NULL,
        mediatype TEXT NOT NULL, 
		thumbnailPath TEXT, 
		width INTEGER, 
		height INTEGER); 
		CREATE TABLE IF NOT EXISTS videoDuration( videoId INTEGER NOT NULL,
		duration INTEGER NOT NULL,
		FOREIGN KEY (videoId) REFERENCES media(id) ON DELETE CASCADE);`

	_, createErr := db.Exec(createString)

	if createErr != nil {
		log.Fatalf("Can not create table because %s", createErr)
	}

	PRAGMA_foreign_keys_String := "PRAGMA foreign_keys=ON"
	_, PRAGMAErr := db.Exec(PRAGMA_foreign_keys_String)

	if PRAGMAErr != nil {
		log.Fatalf("Can not turn on Foreign key support in database")
	}

	return db
}
