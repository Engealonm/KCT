package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
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

	// Check credentials (this is just an example; never hardcode credentials)
	if loginReq.Username == "admin" && loginReq.Password == "password" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"success": "true"})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Invalid credentials"})
	}
}

func main() {
	// Настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir("./static"))) // Обслуживание статических файлов
	http.HandleFunc("/api", apiHandler)                     // API для обработки POST-запросов
	http.HandleFunc("/login", loginHandler)
	// Запуск сервера
	port := "8080"
	fmt.Printf("Сервер запущен на порту %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
