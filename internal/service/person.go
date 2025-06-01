package service

import (
	"github.com/k1lls3x/person-service/internal/entity"
	"github.com/k1lls3x/person-service/internal/repository"
	"github.com/rs/zerolog/log"
	"fmt"
)

func deref(s *string) string {
	if s != nil {
		return *s
	}
	return "null"
}

func derefInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

func CreatePerson(person *entity.Person) error {

	tx, err := repository.DB.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if r:= recover(); err != nil {
			tx.Rollback()
			log.Error().Interface("panic", r).Msg("Rolled back transaction")
			panic(r)
		}
	}()

		log.Debug().
		Str("name", person.Name).
		Str("surname", person.Surname).
		Msg("Starting person enrichment")

		if err := enrichFromAPI(person); err != nil {
			log.Error().Err(err).Msg("Failed to enrich person from API")
			tx.Rollback()
			return fmt.Errorf("failed to enrich person: %w", err)
		}

		log.Info().
		Str("name", person.Name).
		Str("surname", person.Surname).
		Str("gender", deref(person.Gender)).
		Int("age", derefInt(person.Age)).
		Str("nationality", deref(person.Nationality)).
		Msg("Person enriched successfully")

		log.Debug().Msg("Inserting person into database")
		query:= `
			INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
					VALUES (:name, :surname, :patronymic, :age, :gender, :nationality)
		`
		_, err = repository.DB.NamedExec(query, person)

		if err != nil {
			log.Error().
							Err(err).
							Str("name", person.Name).
							Str("surname", person.Surname).
							Msg("Failed to insert person")
							tx.Rollback()
							return fmt.Errorf("failed to insert person: %w", err)
		}

		if err:= tx.Commit(); err != nil {
			log.Error().Err(err).Msg("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Info().
					Str("name", person.Name).
					Str("surname", person.Surname).
					Msg("Person created successfully")

    return nil
}
