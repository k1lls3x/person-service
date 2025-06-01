package service

import (
	"github.com/k1lls3x/person-service/internal/entity"
	"github.com/k1lls3x/person-service/internal/repository"
	"github.com/rs/zerolog/log"
	"fmt"
	"time"

	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := repository.DB.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return fmt.Errorf("❌ failed to start transaction: %w", err)
	}

	defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        log.Error().Interface("panic", r).Msg("❌ Rolled back transaction")
        panic(r)
    }
	}()

		log.Debug().
		Str("name", person.Name).
		Str("surname", person.Surname).
		Msg("Starting person enrichment")

		if err := enrichFromAPI(ctx,person); err != nil {
			log.Error().Err(err).Msg("❌ Failed to enrich person from API")
			tx.Rollback()
			return fmt.Errorf("❌ failed to enrich person: %w", err)
		}

		log.Info().
		Str("name", person.Name).
		Str("surname", person.Surname).
		Str("gender", deref(person.Gender)).
		Int("age", derefInt(person.Age)).
		Str("nationality", deref(person.Nationality)).
		Msg("Person enriched successfully")

		log.Debug().Msg("Inserting person into database")

		query := `
				INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
				VALUES (:name, :surname, :patronymic, :age, :gender, :nationality)
				RETURNING id, created_at
		`

		rows, err := repository.DB.NamedQuery(query, person)
		if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert person: %w", err)
		}
		defer rows.Close()

		if !rows.Next() {
				tx.Rollback()
				return fmt.Errorf("no rows returned")
		}

		if err := rows.StructScan(person); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to scan returned values: %w", err)
		}

		if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit failed: %w", err)
		}
    return nil
}

func DeletePersonById(id int) error {
	log.Debug().
		Int("id", id).
		Msg("Starting deleting person by id")

	query := `
		DELETE FROM persons WHERE id = $1
	`

	result, err := repository.DB.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to delete person by id")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to get rows affected after delete")
		return err
	}

	if rowsAffected == 0 {
		log.Warn().
			Int("id", id).
			Msg("No person found to delete")
		return fmt.Errorf("person with id %d not found", id)
	}

	log.Info().
		Int("id", id).
		Msg("Successfully deleted person")

	return nil
}
