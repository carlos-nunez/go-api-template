package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/carlos-nunez/go-api-template/model"
	"github.com/carlos-nunez/go-api-template/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var DEPLOY_URL = os.Getenv("DEPLOY_URL")
var DEPLOY_KEY = os.Getenv("DEPLOY_KEY")

func (a API) CreateWebsocketServer(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)

	var server model.WebsocketServer
	err = a.marshallBody(&server, w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	apiToken, err := services.GenerateApiToken()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	server.ApiToken = apiToken
	server.UserEmail = user.Email
	server.Status = "Creating..."

	var foundServer model.WebsocketServer
	err = a.mdb.Collection("servers").FindOne(a.ctx, bson.D{{Key: "uuid", Value: server.UUID}}).Decode(&foundServer)

	if err == nil {
		http.Error(w, "Server with this unique ID already exists. Please try another one.", 500)
		return
	}

	result, err := a.mdb.Collection("servers").InsertOne(a.ctx, server)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	server.ID = result.InsertedID.(primitive.ObjectID)

	a.sendDeployRequest(server)

	js, err := json.Marshal(server)

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) DestroyWebsocketServer(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	var foundServer model.WebsocketServer
	err = a.mdb.Collection("servers").FindOne(a.ctx, bson.D{{Key: "uuid", Value: uuid}}).Decode(&foundServer)

	if err != nil {
		http.Error(w, "Server not found.", 500)
		return
	}

	if foundServer.UserEmail != user.Email {
		http.Error(w, "Server not yours, can't destroy.", 500)
		return
	}

	a.sendDestroyRequest(foundServer)

	js, err := json.Marshal(foundServer)

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) sendDeployRequest(server model.WebsocketServer) (*http.Response, error) {
	payload, err := json.Marshal(server)
	fmt.Println("Creating a ws server.")

	req, err := http.NewRequest("POST", DEPLOY_URL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DEPLOY_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Response Status:", resp.Status)

	return resp, nil
}

func (a API) sendDestroyRequest(server model.WebsocketServer) (*http.Response, error) {
	payload, err := json.Marshal(server)

	req, err := http.NewRequest("DELETE", DEPLOY_URL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DEPLOY_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)

	return resp, nil
}

func (a API) GetWSServerByUUID(uuid string) (model.WebsocketServer, error) {
	var server model.WebsocketServer
	err := a.mdb.Collection("servers").FindOne(a.ctx, bson.D{{Key: "uuid", Value: uuid}}).Decode(&server)

	if err != nil {
		return server, err
	}

	return server, nil
}

func (a API) GetWebsocketToken(token string) string {
	var server model.WebsocketServer
	err := a.mdb.Collection("servers").FindOne(a.ctx, bson.D{{Key: "token", Value: token}}).Decode(&server)

	if err != nil {
		return ""
	}

	return server.ApiToken
}

func (a API) FetchUserWebsocketServers(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserByToken(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var servers []model.WebsocketServer

	cur, err := a.mdb.Collection("servers").Find(a.ctx, bson.D{{Key: "user_email", Value: user.Email}})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cur.Close(a.ctx)

	if err = cur.All(a.ctx, &servers); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	js, _ := json.Marshal(servers)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
