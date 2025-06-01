package service

import (
	"context"
	"time"
	"github.com/k1lls3x/person-service/internal/entity"
	"github.com/k1lls3x/person-service/pkg"
		"github.com/rs/zerolog/log"
)

func enrichFromAPI(parentCtx context.Context, person *entity.Person) error {
	ctx, cancel := context.WithTimeout(parentCtx, 3*time.Second)
	defer cancel()

	type result struct {
		age         *int
		gender      *string
		nationality *string
		err         error
	}

	log.Debug().
		Str("name", person.Name).
		Msg("Starting enrichment from APIs")

	ch := make(chan result, 3)

	go func() {
		age, err := pkg.FetchAge(ctx, person.Name)
		if err != nil {
			log.Error().Err(err).Str("name", person.Name).Msg("Failed to fetch age")
		}
		ch <- result{age: age, err: err}
	}()

	go func() {
		nat, err := pkg.FetchNationality(ctx, person.Name)
		if err != nil {
			log.Error().Err(err).Str("name", person.Name).Msg("Failed to fetch nationality")
		}
		ch <- result{nationality: nat, err: err}
	}()

	go func() {
		gender, err := pkg.FetchGender(ctx, person.Name)
		if err != nil {
			log.Error().Err(err).Str("name", person.Name).Msg("Failed to fetch gender")
		}
		ch <- result{gender: gender, err: err}
	}()

	var finalError error

	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			log.Error().
				Str("name", person.Name).
				Msg("Enrichment context deadline exceeded")
			return ctx.Err()

		case res := <-ch:
			if res.err != nil {
				finalError = res.err
			}
			if res.age != nil {
				person.Age = res.age
				log.Debug().Int("age", *res.age).Str("name", person.Name).Msg("Age enriched")
			}
			if res.gender != nil {
				person.Gender = res.gender
				log.Debug().Str("gender", *res.gender).Str("name", person.Name).Msg("Gender enriched")
			}
			if res.nationality != nil {
				person.Nationality = res.nationality
				log.Debug().Str("nationality", *res.nationality).Str("name", person.Name).Msg("Nationality enriched")
			}
		}
	}

	if finalError != nil {
		log.Warn().Err(finalError).Str("name", person.Name).Msg("Enrichment completed with errors")
	} else {
		log.Info().Str("name", person.Name).Msg("Enrichment completed successfully")
	}

	return finalError
}
