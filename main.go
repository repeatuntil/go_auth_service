package main

import (
	"auth_service/database"
	"auth_service/handlers"
	"auth_service/logger"
	"os"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	logger.DoConsoleLog()
}

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
	http.ListenAndServe(":" + port, handlers.AccessLogMiddleware(router))
}
