package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"

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

func Autherize(c *gin.Context) error {
	//Need to look at the session token then look at the csrf token and finally autherize user

	cookie, cError := c.Request.Cookie("session_token")

	if cError != nil || cookie.Value == "" {
		return errors.New("Unauthorized")
	}

	row := db.QueryRow("SELECT token, csrfToken FROM session WHERE token=?", cookie.Value)

	var session Session

	err := row.Scan(&session.session, &session.csrf)

	if err != nil || session.session != cookie.Value {
		return errors.New("Unauthorized")
	}

	csrf := c.Request.Header.Get("X-CSRF-Token")

	if csrf == "" || csrf != session.csrf {
		return errors.New("Unauthorized")
	}

	return nil
}
