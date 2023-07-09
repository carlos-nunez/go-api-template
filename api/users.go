package api

import (
	"encoding/json"
	"net/http"

	"strings"

	"github.com/carlos-nunez/go-api-template/model"
	"github.com/carlos-nunez/go-api-template/services"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (a API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User

	err := a.marshallBody(&user, w, r)

	if user.Email == "" {
		http.Error(w, "Please enter an email.", 500)
		return
	}
	if user.Password == "" {
		http.Error(w, "Please enter a password", 500)
		return
	}
	if user.FullName == "" {
		http.Error(w, "Please enter a full name.", 500)
		return
	}

	user.Password = services.HashPassword(user.Password)
	ws_token, _ := services.GenerateWSToken(32)
	token, err := services.GenerateToken(user.Email)
	user.WS_Token = ws_token
	user.Token = token
	user.Rank = "User"

	result, err := a.mdb.Collection("users").InsertOne(a.ctx, user)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	js, err := json.Marshal(user)

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) FetchUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	var person model.User
	err := a.mdb.Collection("users").FindOne(a.ctx, bson.D{{Key: "email", Value: email}}).Decode(&person)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	js, _ := json.Marshal(person)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) LoginUser(w http.ResponseWriter, r *http.Request) {
	var userRequest model.User
	a.marshallBody(&userRequest, w, r)

	var userFound model.User

	err := a.mdb.Collection("users").FindOne(a.ctx, bson.D{{Key: "email", Value: userRequest.Email}}).Decode(&userFound)

	if err != nil {
		http.Error(w, "Email not found", 500)
		return
	}

	passwordCorrect := services.ComparePassword(userFound.Password, userRequest.Password)

	if !passwordCorrect {
		http.Error(w, "Email or password incorrect", 500)
		return
	}

	token, err := services.GenerateToken(userFound.Email)

	if err != nil {
		http.Error(w, "Issue occurred with token", 500)
		return
	}

	update := bson.M{"$set": bson.M{"token": token}}
	_, err = a.mdb.Collection("users").UpdateOne(a.ctx, bson.D{{Key: "_id", Value: userFound.ID}}, update)
	if err != nil {
		http.Error(w, "Failed to save token", 500)
		return
	}

	userFound.Token = token

	js, err := json.Marshal(userFound)
	if err != nil {
		http.Error(w, "Failed to marshal user data", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) getUserByToken(w http.ResponseWriter, r *http.Request) (model.User, error) {
	prefix := "Bearer "
	authHeader := r.Header.Get("Authorization")
	reqToken := strings.TrimPrefix(authHeader, prefix)

	var person model.User
	err := a.mdb.Collection("users").FindOne(a.ctx, bson.D{{Key: "token", Value: reqToken}}).Decode(&person)

	if err != nil {
		return model.User{}, err
	}

	return person, nil
}

func (a API) GetUserByWSToken(token string) (model.User, error) {
	var person model.User
	err := a.mdb.Collection("users").FindOne(a.ctx, bson.D{{Key: "ws_token", Value: token}}).Decode(&person)

	if err != nil {
		return model.User{}, err
	}

	return person, nil
}

func (a API) FetchUserByToken(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	js, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) RegenerateWSToken(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	token, err := services.GenerateWSToken(32)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	update := bson.M{"$set": bson.M{"ws_token": token}}
	_, err = a.mdb.Collection("users").UpdateOne(a.ctx, bson.D{{Key: "_id", Value: user.ID}}, update)
	if err != nil {
		http.Error(w, "Failed to save token", 500)
		return
	}

	js, _ := json.Marshal(token)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
