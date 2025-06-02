package service

import (
	"github.com/k1lls3x/person-service/internal/entity"

	"github.com/rs/zerolog/log"
	"fmt"
	"time"
	"github.com/jmoiron/sqlx"
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

type PersonService struct {
	db *sqlx.DB
}

func NewPersonService(db *sqlx.DB) *PersonService {
	return &PersonService{db: db}
}

// CreatePerson godoc
// @Summary Создать нового человека
// @Description Добавляет человека с обогащением через внешние API
// @Tags persons
// @Accept json
// @Produce json
// @Param person body entity.Person true "Персона"
// @Success 201 {object} entity.Person
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "server error"
// @Router /api/persons [post]
func (s *PersonService) CreatePerson(person *entity.Person) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.Beginx()
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

		rows, err := tx.NamedQuery(query, person)
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

func (s *PersonService) DeletePersonById(id int) error {
	log.Debug().
		Int("id", id).
		Msg("Starting deleting person by id")

	query := `
		DELETE FROM persons WHERE id = $1
	`

	result, err := s.db.Exec(query, id)
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
		return fmt.Errorf("❌ person with id %d not found", id)
	}

	log.Info().
		Int("id", id).
		Msg("✅ Successfully deleted person")

	return nil
}

// GetPersons godoc
// @Summary Получить список людей с фильтрами и пагинацией
// @Tags persons
// @Accept json
// @Produce json
// @Param name query string false "Имя"
// @Param surname query string false "Фамилия"
// @Param gender query string false "Пол"
// @Param nationality query string false "Национальность"
// @Param minAge query int false "Мин. возраст"
// @Param maxAge query int false "Макс. возраст"
// @Param page query int false "Страница"
// @Param pageSize query int false "Размер страницы"
// @Success 200 {array} entity.Person
// @Router /api/persons [get]
func (s *PersonService) GetPersons(filter entity.PersonFilter) ([]entity.Person, error){
	log.Debug().Msg("Fetching persons with filters")

	type condition struct {
    field string
    op    string
    value interface{}
	}

	conditions := []condition{}
	if filter.Name != nil {
		conditions = append(conditions, condition{"name", "ILIKE", "%" + *filter.Name + "%"})
	}

	if filter.Surname != nil {
		conditions = append(conditions, condition{"surname", "ILIKE", "%" + *filter.Surname + "%"})
	}

	if filter.Gender != nil {
		conditions = append(conditions, condition{"gender", "=", *filter.Gender})
	}

	if filter.Nationality != nil {
		conditions = append(conditions, condition{"nationality", "=", *filter.Nationality})
	}

	if filter.MinAge != nil {
		conditions = append(conditions, condition{"age", ">=", *filter.MinAge})
	}

	if filter.MaxAge != nil {
		conditions = append(conditions, condition{"age", "<=", *filter.MaxAge})
	}
	query := `SELECT * FROM persons WHERE 1=1`
	args:=[]interface{}{}
	for i,cond := range conditions {
		query += fmt.Sprintf(" AND %s %s $%d", cond.field,cond.op,i + 1)
		args = append(args, cond.value)
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	offset := (filter.Page - 1) * filter.PageSize

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, filter.PageSize, offset)

	log.Debug().
		Str("query", query).
		Int("args_len", len(args)).
		Msg("Final SQL query")

		rows, err := s.db.Queryx(query, args...)
		if err != nil {
			log.Error().Err(err).Msg("❌ Failed to query persons")
			return nil, err
		}
		defer rows.Close()

		var persons []entity.Person

		for rows.Next() {
			var person entity.Person
			if err := rows.StructScan(&person); err != nil {
				log.Error().Err(err).Msg("❌ Failed to scan person row")
				return nil, err
			}
			persons = append(persons, person)
		}
		log.Info().Int("count", len(persons)).Msg("✅ Persons fetched successfully")
		return persons, nil
}

func (s *PersonService) UpdatePerson(id int, updatedPerson *entity.Person) error {
	log.Debug().Msg("Change person starting")
	ctx, cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	tx, err := s.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return err
	}

	defer func(){
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error().Interface("panic", r).Msg("Rolled back transaction due to panic")
			panic(r)
		}
	}()

	if err := enrichFromAPI(ctx, updatedPerson); err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to enrich person")
		return err
	}
	query := `
	UPDATE persons
		SET
			name = :name,
			surname = :surname,
			patronymic = :patronymic,
			age = :age,
			gender = :gender,
			nationality = :nationality,
			updated_at = NOW()
		WHERE id = :id;
	`
	updatedPerson.ID = id
	_, err = tx.NamedExecContext(ctx,query,updatedPerson)
	if err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to update person")
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Info().
	Int("id", id).
	Str("name", updatedPerson.Name).
	Str("surname", updatedPerson.Surname).
	Msg("Person updated successfully")
 return nil
}
