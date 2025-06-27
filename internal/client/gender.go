package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

func (c *APIClient) FetchGender(ctx context.Context, name string) (*string, error) {
	apiURL := c.GenderURL + "?name=" + url.PathEscape(name)

	log.Info().Str("name", name).Msg("Fetching gender from API")
	log.Debug().Str("url", apiURL).Msg("Sending request to external API")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		log.Error().Err(err).Str("url", apiURL).Msg("Failed to create HTTP request")
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
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
