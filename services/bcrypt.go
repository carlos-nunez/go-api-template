package services

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	if err != nil {
		fmt.Println(err)
	}
	return string(bytes)
}

func ComparePassword(hashedUserPassword string, attemptedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedUserPassword), []byte(attemptedPassword))
	if err != nil {
		return false
	}
	return true
}
