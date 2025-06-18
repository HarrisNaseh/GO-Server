package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// func hashPassword(password string) (string, error) {

// 	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
// 	return string(bytes), err
// }

func checkPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	return err == nil
}

func generateToken(length int) string {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

func login(c *gin.Context) {

	var loginData LoginRequest
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.String(http.StatusBadRequest, "Invalid login data")
		return
	}

	// username := c.Request.FormValue("username")
	// plainPassword := c.Request.FormValue("password")

	row := db.QueryRow("SELECT * FROM user WHERE username=?", loginData.Username)

	var user User
	err := row.Scan(&user.id, &user.username, &user.password)

	if err != nil || !checkPassword(loginData.Password, user.password) {
		c.String(http.StatusUnauthorized, "Wrong Username or password")
		return
	}

	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	maxAge := 1 * 48 * 60 * 60
	var currTime = time.Now()
	_, dbErr := db.Exec("INSERT INTO session (token, csrfToken, username, createdAt) VALUES (?,?,?,?)", sessionToken, csrfToken, user.username, currTime)

	if dbErr != nil {
		c.String(http.StatusInternalServerError, "Problem With auth")
		fmt.Printf("%v", err)
		return
	}

	//set sucure argument to true when using https
	c.SetCookie("session_token", sessionToken, maxAge, "/", "", false, true)
	// c.SetCookie("CSRF-Token", csrfToken, 1, "/", "", false, false)
	c.JSON(http.StatusOK, gin.H{
		"user":       gin.H{"username": user.username},
		"csrf_token": csrfToken})

}

func logout(c *gin.Context) {
	//delete session from database
	tokenCookie, _ := c.Request.Cookie("session_token")

	token, err := url.QueryUnescape(tokenCookie.Value)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid session token")
		return
	}

	_, dbErr := db.Exec("DELETE FROM session WHERE token=?", token)
	if dbErr != nil {
		c.String(http.StatusInternalServerError, "Could not sign out")
		return
	}

	c.SetCookie("session_token", "", 0, "", "", false, true)
	// c.SetCookie("csrf_token", "", 0, "/", "", false, false)

	c.String(http.StatusOK, "Signed out successfully")
	fmt.Println("Signed out successfully")
}

func checkAuthStatus(c *gin.Context) {
	tokenCookie, _ := c.Request.Cookie("session_token")

	if tokenCookie == nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false, "error": "Not authenticated"})
		return
	}

	token, err := url.QueryUnescape(tokenCookie.Value)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid session token")
		return
	}

	var session Session
	row := db.QueryRow("SELECT token, username, createdAt FROM session WHERE token=?", token)
	err = row.Scan(&session.session, &session.user, &session.createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"authenticated": false, "error": "No session"})
			return
		}
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	if time.Since(session.createdAt) > 48*time.Hour {
		c.JSON(http.StatusOK, gin.H{"authenticated": false, "error": "Session expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"authenticated": true,
		"user": gin.H{"username": session.user},
	})

}
