package pkg

import (
	"encoding/json"
	"net/http"
	"net/url"
	"context"
	"github.com/rs/zerolog/log"
)

func FetchAge(ctx context.Context, name string) (*int, error) {
	apiURL := "https://api.agify.io/?name=" + url.PathEscape(name)

	log.Info().Str("name", name).Msg("Fetching age from API")
	log.Debug().Str("url", apiURL).Msg("Sending request to external API")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		log.Error().Err(err).Str("url", apiURL).Msg("Failed to create HTTP request")
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", apiURL).Msg("Failed to send request to external API")
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Age *int `json:"age"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error().Err(err).Str("url", apiURL).Msg("Failed to decode response from API")
		return nil, err
	}

	if result.Age != nil {
		log.Info().Str("name", name).Int("age", *result.Age).Msg("Successfully fetched age from API")
} else {
		log.Info().Str("name", name).Msg("No age returned from API")
}
return result.Age, nil
}
