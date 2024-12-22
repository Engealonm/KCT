package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Token struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Amount    int                `bson:"amount" json:"amount"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func connectDB() (*mongo.Client, *mongo.Collection) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")
	collection := client.Database("mydb").Collection("tokens")

	return client, collection
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Status:  "fail",
			Message: "Method not allowed",
		})
		return
	}

	var req map[string]string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req["message"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Status:  "fail",
			Message: "Invalid JSON payload",
		})
		return
	}

	fmt.Printf("Message from client: %s\n", req["message"])

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Status:  "success",
		Message: "Data received successfully",
	})
}

func createTokenHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var token Token
		err := json.NewDecoder(r.Body).Decode(&token)
		if err != nil || token.Name == "" || token.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Invalid data",
			})
			return
		}

		token.CreatedAt = time.Now()
		res, err := collection.InsertOne(context.TODO(), token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Error creating token",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Token created",
			Data:    res.InsertedID,
		})
	}
}

func listTokensHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var tokens []Token
		cursor, err := collection.Find(context.TODO(), bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Error fetching tokens",
			})
			return
		}
		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var token Token
			if err := cursor.Decode(&token); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(Response{
					Status:  "fail",
					Message: "Error decoding token",
				})
				return
			}
			tokens = append(tokens, token)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Tokens fetched",
			Data:    tokens,
		})
	}
}

func deleteTokenHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "ID is required",
			})
			return
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Invalid ID format",
			})
			return
		}

		_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Error deleting token",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Token deleted",
		})
	}
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if loginReq.Username == "admin" && loginReq.Password == "password" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"success": "true"})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Invalid credentials"})
	}
}
func main() {
	client, collection := connectDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
		fmt.Println("Disconnected from MongoDB.")
	}()
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/tokens/create", createTokenHandler(collection))
	http.HandleFunc("/tokens", listTokensHandler(collection))
	http.HandleFunc("/tokens/delete", deleteTokenHandler(collection))

	port := "8080"
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
