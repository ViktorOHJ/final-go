package main

import (
	"go1f/pkg/api"
	"go1f/pkg/db"

	"go1f/pkg/server"
	"log"
)

func main() {
	err := db.Init("scheduler.db")
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.DB.Close()
	log.Println("База данных инициализирована и подключена.")
	api.Init()
	log.Println("Попытка запуска сервера...")
	err = server.Run()
	if err != nil {
		log.Printf("Ошибка запуска сервера: %v", err)
		return
	}
	log.Println("Сервер запущен на порту 7540.")
}
