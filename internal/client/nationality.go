package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

func ptrString(str string) *string {
	return &str
}

func (c *APIClient) FetchNationality(ctx context.Context, name string) (*string, error) {
	apiURL := c.NationalityURL + "?name=" + url.PathEscape(name)

	log.Info().Str("name", name).Msg("Fetching nationality from API")
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
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error().Err(err).Str("url", apiURL).Msg("Failed to decode response from API")
		return nil, err
	}

	if len(result.Country) == 0 {
		log.Info().Str("name", name).Msg("No nationality data found")
		return nil, nil
	}

	nationality := result.Country[0].CountryID

	log.Info().
		Str("name", name).
		Str("nationality", nationality).
		Float64("probability", result.Country[0].Probability).
		Msg("Successfully fetched nationality from API")

	return ptrString(nationality), nil
}
