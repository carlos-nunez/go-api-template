package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/carlos-nunez/go-api-template/model"
	"github.com/carlos-nunez/go-api-template/services"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (a API) FetchSupportTickets(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var tickets []model.SupportTicket
	cur, err := a.mdb.Collection("tickets").Find(a.ctx, bson.D{{Key: "user_email", Value: user.Email}})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cur.Close(a.ctx)
	if err = cur.All(a.ctx, &tickets); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	js, _ := json.Marshal(tickets)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) FetchAllSupportTickets(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil || user.Rank != "Admin" {
		http.Error(w, err.Error(), 500)
		return
	}

	var tickets []model.SupportTicket
	cur, err := a.mdb.Collection("tickets").Find(a.ctx, bson.D{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cur.Close(a.ctx)
	if err = cur.All(a.ctx, &tickets); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	js, _ := json.Marshal(tickets)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) FetchSupportTicket(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	objID, err := primitive.ObjectIDFromHex(id)

	var ticket model.SupportTicket
	err = a.mdb.Collection("tickets").FindOne(a.ctx, bson.D{{Key: "_id", Value: objID}}).Decode(&ticket)
	if err != nil {
		http.Error(w, "Ticket not found.", 500)
		return
	}

	if ticket.UserEmail != user.Email || user.Rank != "Admin" {
		http.Error(w, "Unable to retrive ticket.", 500)
		return
	}

	js, _ := json.Marshal(ticket)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) CreateSupportTicket(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var ticket model.SupportTicket
	err = a.marshallBody(&ticket, w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if ticket.Title == "" || ticket.Description == "" || ticket.Priority == "" {
		http.Error(w, "Missing required ticket fields.", 500)
		return
	}
	ticket.UserEmail = user.Email
	ticket.UserName = user.FullName
	ticket.CreatedAt = time.Now()
	ticket.Status = "Open"
	result, err := a.mdb.Collection("tickets").InsertOne(a.ctx, ticket)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	ticket.ID = result.InsertedID.(primitive.ObjectID)
	uuid := services.GenerateUniqueStringFromObjectID(ticket.ID)
	ticket.Number = uuid

	update := bson.M{"$set": bson.M{"number": ticket.Number}}
	a.mdb.Collection("tickets").UpdateOne(a.ctx, bson.D{{Key: "_id", Value: ticket.ID}}, update)

	js, err := json.Marshal(ticket)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) AddSupportReply(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, "User not found.", 500)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	objID, err := primitive.ObjectIDFromHex(id)

	var reply model.SupportTicketReply
	err = a.marshallBody(&reply, w, r)
	reply.UserEmail = user.Email
	reply.UserName = user.FullName
	reply.CreatedAt = time.Now()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var ticket model.SupportTicket
	err = a.mdb.Collection("tickets").FindOne(a.ctx, bson.D{{Key: "_id", Value: objID}}).Decode(&ticket)
	if err != nil {
		http.Error(w, "Ticket not found.", 500)
		return
	}
	if ticket.Replies == nil {
		ticket.Replies = []model.SupportTicketReply{reply}
	} else {
		ticket.Replies = append(ticket.Replies, reply)
	}
	update := bson.M{"$set": bson.M{
		"replies": ticket.Replies,
	}}
	result, err := a.mdb.Collection("tickets").UpdateOne(a.ctx, bson.D{{Key: "_id", Value: objID}}, update)
	if result.MatchedCount == 0 {
		http.Error(w, "No ticket found with provided ID.", 404)
		return
	}

	js, err := json.Marshal(ticket)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	_, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, "User not found.", 500)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ticket ID.", 500)
		return
	}
	var ticket model.SupportTicket
	err = a.marshallBody(&ticket, w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	update := bson.M{"$set": bson.M{
		"status": ticket.Status,
	}}
	result, err := a.mdb.Collection("tickets").UpdateOne(a.ctx, bson.D{{Key: "_id", Value: objID}}, update)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if result.MatchedCount == 0 {
		http.Error(w, "No ticket found with provided ID.", 404)
		return
	}

	js, err := json.Marshal(ticket)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
