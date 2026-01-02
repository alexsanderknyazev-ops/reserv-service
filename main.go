package main

import (
	"log"
	"reserv-service/database"
	"reserv-service/route"
)

func main() {
	// 1. Сначала инициализируем БД
	database.InitDB()
	log.Println("✅ Database initialized")

	// 2. Получаем роутер (не запускаем сервер)
	r := route.InitRoute()
	log.Println("✅ Router initialized")

	// 3. Запускаем сервер
	log.Println("🚀 Server starting on :8074")
	if err := r.Run(":8074"); err != nil {
		log.Fatal("❌ Server error: ", err)
	}
}
