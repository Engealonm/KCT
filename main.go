package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Response структура для формирования JSON-ответа
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Request структура для парсинга JSON-запроса
type Request struct {
	Message string `json:"message"`
}

// handler функция для обработки запросов
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Status:  "fail",
			Message: "Метод не поддерживается",
		})
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Status:  "fail",
			Message: "Некорректное JSON-сообщение",
		})
		return
	}

	// Логируем полученное сообщение
	fmt.Printf("Сообщение от клиента: %s\n", req.Message)

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Status:  "success",
		Message: "Данные успешно приняты",
	})
}

func main() {
	// Установка маршрутов
	http.HandleFunc("/", handler)

	// Запуск сервера
	port := "8080"
	fmt.Printf("Сервер запущен на порту %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
