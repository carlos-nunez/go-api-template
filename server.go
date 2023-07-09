package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	API "github.com/carlos-nunez/go-api-template/api"
	"github.com/carlos-nunez/go-api-template/middleware"
	ws "github.com/carlos-nunez/go-api-template/ws"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	api    = API.NewAPI()
	mdb    mongo.Database
	router *mux.Router
	ctx    context.Context
)

func setupAPI() {
	fetch := router.Methods("GET").PathPrefix("/api").Subrouter()
	update := router.Methods("PUT").PathPrefix("/api").Subrouter()
	create := router.Methods("POST").PathPrefix("/api").Subrouter()
	delete := router.Methods("DELETE").PathPrefix("/api").Subrouter()

	create.HandleFunc("/users", api.CreateUser)
	create.HandleFunc("/users/login", api.LoginUser)
	fetch.HandleFunc("/users/regenerateToken", middleware.Auth(api.RegenerateWSToken))
	fetch.HandleFunc("/users/current", middleware.Auth(api.FetchUserByToken))

	fetch.HandleFunc("/servers", middleware.Auth(api.FetchUserWebsocketServers))
	create.HandleFunc("/servers", middleware.Auth(api.CreateWebsocketServer))
	delete.HandleFunc("/servers/{uuid}", middleware.Auth(api.DestroyWebsocketServer))

	fetch.HandleFunc("/tickets", middleware.Auth(api.FetchSupportTickets))
	fetch.HandleFunc("/tickets/all", middleware.Auth(api.FetchAllSupportTickets))
	create.HandleFunc("/tickets", middleware.Auth(api.CreateSupportTicket))
	update.HandleFunc("/tickets/{id}/reply", middleware.Auth(api.AddSupportReply))
	update.HandleFunc("/tickets/{id}/status", middleware.Auth(api.UpdateTicketStatus))

	fmt.Println("Finished Setting Up API")
}

func setupMongo() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No Env File")
	}

	uri := os.Getenv("MONGO_URI")

	if len(uri) == 0 {
		panic("No MongoURI")
	}

	ctx = context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	database := os.Getenv("PRODUCT")
	db := client.Database(database)
	mdb = *db

	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("Successfully Connected to MongoDB")
}

func setupIndexes(mdb mongo.Database, ctx context.Context) {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	websocketIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	name, _ := mdb.Collection("users").Indexes().CreateOne(ctx, indexModel)
	fmt.Println("Name of Index Created:", name)

	name2, err := mdb.Collection("servers").Indexes().CreateOne(ctx, websocketIndexModel)
	if err != nil {
		fmt.Println("Error creating index:", err)
	} else {
		fmt.Println("Name of Index Created:", name2)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("r.URL.Path: " + r.URL.Path)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func setupWS() {
	root := router.Methods("GET").Subrouter()
	root.HandleFunc("/", serveHome)
	hub := ws.NewHub(api)
	go hub.Run()
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})
}

func main() {
	router = mux.NewRouter()

	setupMongo()
	api.Initialize(mdb, ctx)
	setupAPI()
	setupWS()
	setupIndexes(mdb, ctx)

	corsOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	corsHandler := handlers.CORS(corsOrigins, corsMethods, corsHeaders)(router)

	http.ListenAndServe(":5000", corsHandler)
}
