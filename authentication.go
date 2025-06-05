package main

import (
	"crypto/rand"
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
	_, dbErr := db.Exec("INSERT INTO session (token, csrfToken, userId, createdAt) VALUES (?,?,?,?)", sessionToken, csrfToken, user.id, currTime)

	if dbErr != nil {
		c.String(http.StatusInternalServerError, "Problem With auth")
		fmt.Printf("%v", err)
		return
	}

	//set sucure argument to true when using https
	c.SetCookie("session_token", sessionToken, maxAge, "/", "", false, true)
	c.SetCookie("csrf_token", csrfToken, maxAge, "/", "", false, false)

}

func logout(c *gin.Context) {
	//autherize the request

	// if err := Autherize(c); err != nil {
	// 	c.String(http.StatusUnauthorized, "Unatherized access to this route. Sign in")
	// 	return
	// }
	// cookie, _ := c.Request.Cookie("session_token")
	// sessionToken := cookie.Value

	// _, dbErr := db.Exec("DELECT FROM session WHERE token=?", sessionToken)

	// if dbErr != nil {
	// 	c.String(http.StatusInternalServerError, "Could not sign out")
	// }

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
	c.SetCookie("csrf_token", "", 0, "/", "", false, false)

	c.String(http.StatusOK, "Signed out successfully")
	fmt.Println("Signed out successfully")
}
