package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebsocketServer struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	UUID      string             `bson:"uuid" json:"uuid"`
	UserEmail string             `bson:"user_email" json:"user_email"`
	CPU       string             `bson:"cpu" json:"cpu"`
	Memory    string             `bson:"memory" json:"memory"`
	Type      string             `bson:"type" json:"type"`
	Status    string             `bson:"status" json:"status"`
	ApiToken  string             `bson:"token" json:"token,omitempty"`
}
