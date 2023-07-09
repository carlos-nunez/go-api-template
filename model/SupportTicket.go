package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SupportTicketReply struct {
	Text      string    `bson:"text" json:"text"`
	UserEmail string    `bson:"user_email" json:"user_email"`
	UserName  string    `bson:"user_name" json:"user_name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type SupportTicket struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"_id,omitempty"`
	Number      string               `bson:"number" json:"number"`
	Title       string               `bson:"title" json:"title"`
	Priority    string               `bson:"priority" json:"priority"`
	Status      string               `bson:"status" json:"status"`
	Description string               `bson:"description" json:"description"`
	UserEmail   string               `bson:"user_email" json:"user_email"`
	UserName    string               `bson:"user_name" json:"user_name"`
	Product     string               `bson:"product" json:"product"`
	Replies     []SupportTicketReply `bson:"replies" json:"replies"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
}
