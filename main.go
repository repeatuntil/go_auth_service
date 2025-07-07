package main

import (
	"auth_service/database"
	"auth_service/handlers"
	"auth_service/logger"

	_ "auth_service/docs"
	"os"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	logger.DoConsoleLog()
}

// @title API for authentication service
// @version 1.0
// @description This is a documentation for auth service written on golang.
// @host localhost:8080
// @BasePath /

// @contact.email nickita-ananiev@yandex.ru

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Err.Fatalln("can't load .env file")
		return
	}

	conn, err := database.ConfigDataBase(os.Getenv("CONN_BASE"), "postgres", os.Getenv("DB_NAME"), os.Getenv("MIGRATION_DIR"))
	if err != nil {
		logger.Err.Fatalln("Can't configure database connection:", err)
		return
	}

	defer conn.Close()

	router := mux.NewRouter()
	repo := database.NewTokenRepository(conn)
	handler := handlers.NewAuthHandler(router, repo)
	handler.SetUpRoutes()

	port := os.Getenv("SERVE_PORT")
	logger.Debug.Println("all handlers set up now...")
	logger.Debug.Printf("start listening on %s port...\n", port)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	http.ListenAndServe(":" + port, handlers.AccessLogMiddleware(router))
}
