package main

import (
	"log"
	"os"

	"github.com/AgungRilo/bookshelf-app/config"
	"github.com/AgungRilo/bookshelf-app/database"
	"github.com/AgungRilo/bookshelf-app/routes"
	"github.com/joho/godotenv"
)

func main() {
	// .env opsional, variable dari environment tetap bisa dipakai
	godotenv.Load()

	config.ConnectDatabase()
	defer config.DB.Close()

	if err := database.Migrate(config.DB); err != nil {
		log.Fatal("gagal menjalankan migrasi: ", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := routes.SetupRouter()
	r.Run(":" + port)
}
