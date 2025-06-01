package pkg

import (
	"encoding/json"
	"net/http"
	"net/url"
	"context"
	"github.com/rs/zerolog/log"
)

func FetchGender(ctx context.Context, name string) (*string, error) {
	apiURL := "https://api.genderize.io/?name=" + url.PathEscape(name)

	log.Info().Str("name", name).Msg("Fetching gender from API")
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
		Gender string `json:"gender"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error().Err(err).Str("url", apiURL).Msg("Failed to decode response from API")
		return nil, err
	}

	log.Info().Str("name", name).Str("gender", result.Gender).Msg("Successfully fetched gender from API")

	return ptrString(result.Gender), nil
}
