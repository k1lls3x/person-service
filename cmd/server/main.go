// @title Person Service API
// @version 1.0
// @description REST API для сервиса обогащения ФИО возрастом, полом и национальностью
// @host localhost:8888
// @BasePath /
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	 "github.com/swaggo/http-swagger"
	"github.com/k1lls3x/person-service/internal/repository"
	"github.com/k1lls3x/person-service/internal/service"
	"github.com/k1lls3x/person-service/internal/handler"
)

func main() {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found (можно проигнорировать, если переменные уже в окружении)")
	}

	cfg := repository.LoadConfigFromEnv()
	dsn := cfg.DSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ PostgreSQL не отвечает: %v", err)
	}
	log.Println("✅ Подключение к PostgreSQL успешно")


	personService := service.NewPersonService(db)
	h := handler.NewHandler(personService)


	r := chi.NewRouter()
	r.Post("/api/persons", h.CreatePerson)
	r.Get("/api/persons", h.GetPersons)
	r.Put("/api/persons/{id}", h.UpdatePerson)
	r.Delete("/api/persons/{id}", h.DeletePerson)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	addr := ":8888"
	log.Printf("Starting server on %s ...", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
