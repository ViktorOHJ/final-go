package main

import (
	"fmt"
	"log"
	"net/http"

	"go1f/pkg/api"
	"go1f/pkg/db"
)

func main() {
	err := db.Init("scheduler.db")
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.DB.Close()
	log.Println("Подключение к базе данных")
	api.Init()
	port := 7540
	http.Handle("/", http.FileServer(http.Dir("web")))
	log.Println("Запуск сервера на порту 7540.")
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
		return
	}
}
