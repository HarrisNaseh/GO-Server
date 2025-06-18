package main

import (
	_ "image/jpeg"
	_ "image/png"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db = dbInit()

	defer db.Close()

	router := gin.Default()

	router.MaxMultipartMemory = 8 << 20

	router.Use(corsMiddleware())

	router.POST("/login", login)
	router.POST("/logout", logout)
	router.GET("/check-auth", checkAuthStatus)

	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware(db))

	// go router.GET("/video", getVideo)
	authGroup.GET("/media/:id", getMediaById)
	authGroup.GET("/", getAll)
	authGroup.GET("/media/:id/thumbnail", getThumbnailById)

	authGroup.POST("/upload", uploadFiles)

	//detele route
	authGroup.DELETE("/media/:id", deleteById)
	router.Run() //Run this when you want this to run on the network
	// router.Run(":8080")
}
