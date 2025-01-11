package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"net/http"
	"strconv"
	"sync"
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

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

type visitor struct {
	limiter  *time.Ticker
	lastSeen time.Time
}
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

var rateLimiter = newRateLimiter()

func newRateLimiter() *RateLimiter {
	return &RateLimiter{visitors: make(map[string]*visitor)}
}

func (rl *RateLimiter) getVisitor(ip string) *time.Ticker {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := time.NewTicker(time.Second / 100) // 5 запросов в секунду
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := rateLimiter.getVisitor(ip)
		select {
		case <-limiter.C:
			next.ServeHTTP(w, r)
		default:
			logStructured("Rate limiting triggered", map[string]string{"ip": ip})
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}
	})
}
func logStructured(message string, data map[string]string) {
	logEntry := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   message,
		"data":      data,
	}
	logData, _ := json.Marshal(logEntry)
	log.Println(string(logData))
}
func connectDB() (*mongo.Client, *mongo.Collection) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logStructured("Error connecting to MongoDB", map[string]string{"error": err.Error()})
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		logStructured("Error pinging MongoDB", map[string]string{"error": err.Error()})
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	logStructured("Connected to MongoDB", nil)
	collection := client.Database("mydb").Collection("tokens")

	return client, collection
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logStructured("Invalid HTTP method", map[string]string{"method": r.Method})
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Status:  "fail",
			Message: "Method not allowed",
		})
		return
	}

	var req map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req["message"] == "" {
		logStructured("Invalid JSON payload", map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Status:  "fail",
			Message: "Invalid JSON payload",
		})
		return
	}

	logStructured("Valid request received", map[string]string{"message": fmt.Sprintf("%v", req["message"])})
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

		// Get query parameters for pagination, filtering, and sorting
		sortOrder := r.URL.Query().Get("sortOrder")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

		// Default values for pagination
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 10
		}

		// Filters for name and amount
		nameFilter := r.URL.Query().Get("name")
		minAmount, _ := strconv.Atoi(r.URL.Query().Get("minAmount"))
		maxAmount, _ := strconv.Atoi(r.URL.Query().Get("maxAmount"))

		// Build the sort field
		sortField := bson.D{}
		if sortOrder == "asc" {
			sortField = bson.D{{"amount", 1}} // Ascending order by amount
		} else if sortOrder == "desc" {
			sortField = bson.D{{"amount", -1}} // Descending order by amount
		}

		// Create Find options for pagination and sorting
		findOptions := options.Find()
		if len(sortField) > 0 {
			findOptions.SetSort(sortField)
		}
		findOptions.SetSkip(int64((page - 1) * limit)) // Skip for pagination
		findOptions.SetLimit(int64(limit))             // Limit results per page

		// Build the filter object
		filter := bson.M{}
		if nameFilter != "" {
			filter["name"] = bson.M{"$regex": nameFilter, "$options": "i"} // Case-insensitive name search
		}
		if minAmount > 0 {
			filter["amount"] = bson.M{"$gte": minAmount} // Amount greater than or equal to minAmount
		}
		if maxAmount > 0 {
			if _, ok := filter["amount"]; ok {
				filter["amount"].(bson.M)["$lte"] = maxAmount // Amount less than or equal to maxAmount
			} else {
				filter["amount"] = bson.M{"$lte": maxAmount} // Amount less than or equal to maxAmount
			}
		}

		// Query MongoDB with the filter and find options
		var tokens []Token
		cursor, err := collection.Find(context.TODO(), filter, findOptions)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Error fetching tokens",
			})
			return
		}
		defer cursor.Close(context.TODO())

		// Decode each token from the cursor
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

		// Count total number of tokens for pagination metadata
		totalCount, err := collection.CountDocuments(context.TODO(), filter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Error counting tokens",
			})
			return
		}

		// Calculate total number of pages
		totalPages := int(totalCount) / limit
		if totalCount%int64(limit) != 0 {
			totalPages++
		}

		// Return the paginated and filtered result with pagination metadata
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Tokens fetched",
			Data: map[string]interface{}{
				"tokens":      tokens,
				"currentPage": page,
				"totalPages":  totalPages,
				"totalCount":  totalCount,
				"limit":       limit,
			},
		})
	}
}

func updateTokenHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var updateData struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Amount int    `json:"amount"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Invalid data format",
			})
			return
		}

		// Convert string ID to ObjectID
		objID, err := primitive.ObjectIDFromHex(updateData.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Invalid ID format",
			})
			return
		}

		if updateData.Name == "" || updateData.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Name and amount are required and amount must be positive",
			})
			return
		}

		filter := bson.M{"_id": objID}
		update := bson.M{
			"$set": bson.M{
				"name":       updateData.Name,
				"amount":     updateData.Amount,
				"updated_at": time.Now(),
			},
		}

		res, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Error updating token",
			})
			return
		}

		if res.MatchedCount == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{
				Status:  "fail",
				Message: "Token not found",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Token updated successfully",
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
func searchTokenHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
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

		var token Token
		err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&token)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(Response{
					Status:  "fail",
					Message: "Token not found",
				})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(Response{
					Status:  "fail",
					Message: "Error finding token",
				})
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Token found",
			Data:    token,
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
func sendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var emailReq EmailRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&emailReq)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	m := gomail.NewMessage()
	m.SetHeader("From", "miko2005zez@gmail.com") // Replace with your email
	m.SetHeader("To", emailReq.To)
	m.SetHeader("Subject", emailReq.Subject)
	m.SetBody("text/plain", emailReq.Body)
	// Set up the SMTP server configuration.
	d := gomail.NewDialer("smtp.gmail.com", 587, "miko2005zez@gmail.com", "xlsn rbae psfi wswv") // Replace with your SMTP credentials

	// Create a new message.

	// Send the email.
	err = d.DialAndSend(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending email: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func main() {
	client, collection := connectDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			logStructured("Error disconnecting MongoDB", map[string]string{"error": err.Error()})
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
		logStructured("Disconnected from MongoDB", nil)
	}()

	go rateLimiter.cleanupVisitors()

	http.Handle("/", rateLimitMiddleware(http.FileServer(http.Dir("./static"))))
	http.Handle("/api", rateLimitMiddleware(http.HandlerFunc(apiHandler)))
	http.Handle("/tokens/create", rateLimitMiddleware(http.HandlerFunc(createTokenHandler(collection))))
	http.Handle("/tokens", rateLimitMiddleware(http.HandlerFunc(listTokensHandler(collection))))
	http.Handle("/tokens/delete", rateLimitMiddleware(http.HandlerFunc(deleteTokenHandler(collection))))
	http.Handle("/tokens/update", rateLimitMiddleware(http.HandlerFunc(updateTokenHandler(collection))))
	http.Handle("/tokens/search", rateLimitMiddleware(http.HandlerFunc(searchTokenHandler(collection))))
	http.Handle("/login", rateLimitMiddleware(http.HandlerFunc(loginHandler)))
	http.Handle("/sendmail", rateLimitMiddleware(http.HandlerFunc(sendEmail)))

	port := "8080"
	logStructured("Server starting", map[string]string{"port": port})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
