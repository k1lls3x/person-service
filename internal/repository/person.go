package repository

import (
	"github.com/k1lls3x/person-service/internal/entity"
	"github.com/rs/zerolog/log"
	"fmt"
)

// TODO: НЕ ЗАБЫТЬ ОБОГАТИТЬ ДАННЫЕ ИЗ АПИ
func Create(person *entity.Person) error {
	query:= `
		 INSERT INTO persons (name, surname, patronymic, age, gender, nationality, created_at)
        VALUES (:name, :surname, :patronymic, :age, :gender, :nationality, :created_at)
	`

		log.Debug().
		Str("name", person.Name).
		Str("surname", person.Surname).
		Msg("Creating person in database")

	_, err := DB.NamedExec(query, person)

	if err != nil {
    log.Error().
            Err(err).
            Str("name", person.Name).
            Str("surname", person.Surname).
            Msg("Failed to create person")
        return fmt.Errorf("database error: %w", err)
	}

	log.Info().
        Str("name", person.Name).
        Str("surname", person.Surname).
        Msg("Person created successfully")

    return nil
}
