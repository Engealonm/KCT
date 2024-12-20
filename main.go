package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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

func main() {
	// Настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir("./static"))) // Обслуживание статических файлов
	http.HandleFunc("/api", apiHandler)                     // API для обработки POST-запросов

	// Запуск сервера
	port := "8080"
	fmt.Printf("Сервер запущен на порту %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
