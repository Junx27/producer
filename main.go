package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gitlab.com/trisaptono/producer/config"
	"gitlab.com/trisaptono/producer/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		panic(err)
	}
	r := gin.Default()
	config.LoadEnv()

	config.InitializeDatabase()
	routes.ServerRoutes()

	r.Run() // listen and serve on 0.0.0.0:8080
}
