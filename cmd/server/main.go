// @title Person Service API
// @version 1.0
// @description REST API для сервиса обогащения ФИО возрастом, полом и национальностью
// @host localhost:8888
// @BasePath /
package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/k1lls3x/person-service/internal/client"
	"github.com/k1lls3x/person-service/internal/handler"
	"github.com/k1lls3x/person-service/internal/repository"
	"github.com/k1lls3x/person-service/internal/service"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/swaggo/http-swagger"
)

func main() {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Info().Msg("No .env file found (можно проигнорировать, если переменные уже в окружении)")
	}

	cfg := repository.LoadConfigFromEnv()
	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	dsn := cfg.DSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка подключения к PostgreSQL")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL не отвечает")
	}
	log.Info().Msg("Подключение к PostgreSQL успешно")

	apiClient := client.NewAPIClient(cfg.AgeAPIURL, cfg.GenderAPIURL, cfg.NationalityAPIURL)
	personService := service.NewPersonService(db, apiClient)
	h := handler.NewHandler(personService)

	r := chi.NewRouter()
	r.Post("/api/persons", h.CreatePerson)
	r.Get("/api/persons", h.GetPersons)
	r.Put("/api/persons/{id}", h.UpdatePerson)
	r.Delete("/api/persons/{id}", h.DeletePerson)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	addr := ":8888"
	log.Info().Msgf("Starting server on %s ...", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
