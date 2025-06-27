package client

import "net/http"

// APIClient provides methods to call external enrichment services.
type APIClient struct {
	AgeURL         string
	GenderURL      string
	NationalityURL string
	HTTPClient     *http.Client
}

func NewAPIClient(ageURL, genderURL, natURL string) *APIClient {
	return &APIClient{
		AgeURL:         ageURL,
		GenderURL:      genderURL,
		NationalityURL: natURL,
		HTTPClient:     http.DefaultClient,
	}
}
