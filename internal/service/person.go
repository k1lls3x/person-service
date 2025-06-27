package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/k1lls3x/person-service/internal/client"
	"github.com/k1lls3x/person-service/internal/entity"
)

var ErrNotFound = errors.New("person not found")

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
	db        *sqlx.DB
	apiClient *client.APIClient
}

func NewPersonService(db *sqlx.DB, apiClient *client.APIClient) *PersonService {
	return &PersonService{db: db, apiClient: apiClient}
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
func (s *PersonService) CreatePerson(input *entity.CreatePersonInput) (*entity.Person, error) {
	person := &entity.Person{
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return nil, fmt.Errorf("failed to start transaction: %w", err)
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

	if err := enrichFromAPI(ctx, s.apiClient, person); err != nil {
		log.Error().Err(err).Msg("Failed to enrich person from API")
		tx.Rollback()
		return nil, fmt.Errorf("failed to enrich person: %w", err)
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
		return nil, fmt.Errorf("failed to insert person: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		tx.Rollback()
		return nil, fmt.Errorf("no rows returned")
	}

	if err := rows.StructScan(person); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to scan returned values: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	return person, nil
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
		return ErrNotFound
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
func (s *PersonService) GetPersons(filter entity.PersonFilter) ([]entity.Person, error) {
	log.Debug().Msg("Fetching persons with filters")

	qb := squirrel.Select("*").From("persons").PlaceholderFormat(squirrel.Dollar)

	if filter.Name != nil {
		qb = qb.Where(squirrel.ILike{"name": "%" + *filter.Name + "%"})
	}
	if filter.Surname != nil {
		qb = qb.Where(squirrel.ILike{"surname": "%" + *filter.Surname + "%"})
	}
	if filter.Patronymic != nil {
		qb = qb.Where(squirrel.ILike{"patronymic": "%" + *filter.Patronymic + "%"})
	}
	if filter.Gender != nil {
		qb = qb.Where(squirrel.Eq{"gender": *filter.Gender})
	}
	if filter.Nationality != nil {
		qb = qb.Where(squirrel.Eq{"nationality": *filter.Nationality})
	}
	if filter.MinAge != nil {
		qb = qb.Where(squirrel.GtOrEq{"age": *filter.MinAge})
	}
	if filter.MaxAge != nil {
		qb = qb.Where(squirrel.LtOrEq{"age": *filter.MaxAge})
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	offset := (filter.Page - 1) * filter.PageSize

	qb = qb.OrderBy("created_at DESC").Limit(uint64(filter.PageSize)).Offset(uint64(offset))

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, err
	}

	log.Debug().Str("query", query).Int("args_len", len(args)).Msg("Final SQL query")

	rows, err := s.db.Queryx(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query persons")
		return nil, err
	}
	defer rows.Close()

	var persons []entity.Person
	for rows.Next() {
		var person entity.Person
		if err := rows.StructScan(&person); err != nil {
			log.Error().Err(err).Msg("Failed to scan person row")
			return nil, err
		}
		persons = append(persons, person)
	}

	log.Info().Int("count", len(persons)).Msg("Persons fetched successfully")
	return persons, nil
}

func (s *PersonService) UpdatePerson(id int, input *entity.UpdatePersonInput) (*entity.Person, error) {
	updatedPerson := &entity.Person{
		ID:         id,
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
	}
	log.Debug().Msg("Change person starting")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction")
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error().Interface("panic", r).Msg("Rolled back transaction due to panic")
			panic(r)
		}
	}()

	if err := enrichFromAPI(ctx, s.apiClient, updatedPerson); err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to enrich person")
		return nil, err
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
	res, err := tx.NamedExecContext(ctx, query, updatedPerson)
	if err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to update person")
		return nil, err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		tx.Rollback()
		return nil, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Info().
		Int("id", id).
		Str("name", updatedPerson.Name).
		Str("surname", updatedPerson.Surname).
		Msg("Person updated successfully")
	return updatedPerson, nil
}
