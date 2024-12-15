package main

import (
	"log"
	"os"
	"we-credit/routes"

	"github.com/gin-contrib/pprof"
	"github.com/joho/godotenv"
)

// @title Tutree Swagger API
// @version 2.0
// @description This is swagger api for Tutree.
// @BasePath /api
// @schemes http https
func main() {

	err := godotenv.Load()
	if err != nil {

		log.Fatal("Error loading .env file -> ", err)
	}

	//setup routes
	r := routes.SetupRouter()
	pprof.Register(r)
	// running
	r.Run(":" + os.Getenv("PORT"))
}
