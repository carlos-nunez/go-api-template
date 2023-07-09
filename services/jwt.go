package services

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/carlos-nunez/go-api-template/model"
	jwt "github.com/golang-jwt/jwt"
)

var SIGNING_SECRET = os.Getenv("SIGNING_SECRET")

func GenerateToken(email string) (signedToken string, e error) {
	claims := &model.SignedClaims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SIGNING_SECRET))

	if err != nil {
		log.Panic(err)
	}

	return token, err
}

func ValidateToken(signedToken string) (claim *model.SignedClaims, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&model.SignedClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SIGNING_SECRET), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*model.SignedClaims)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}
