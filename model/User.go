package model

import (
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email    string             `bson:"email" json:"email"`
	FullName string             `bson:"full_name" json:"full_name"`
	Password string             `bson:"password" json:"password"`
	Token    string             `bson:"token" json:"token,omitempty"`
	WS_Token string             `bson:"ws_token" json:"ws_token,omitempty"`
	Rank     string             `bson:"rank" json:"rank,omitempty"`
}

type SignedClaims struct {
	Email string
	jwt.StandardClaims
}
