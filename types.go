package main

import (
	"database/sql"
	"time"
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
	".mov":   "video/mp4",
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

type User struct {
	username string
	password string
}

type Session struct {
	session string
	csrf    string
}

var db *sql.DB
