package data

import (
	"fmt"
	"os"
)

// FootballDataAPIKey retrieves the API-Sports.io API key from environment variable.
// The API key must be set via the FOOTBALL_DATA_API_KEY environment variable.
// Get a free API key at https://www.api-sports.io/
func FootballDataAPIKey() (string, error) {
	apiKey := os.Getenv("FOOTBALL_DATA_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("FOOTBALL_DATA_API_KEY environment variable not set")
	}

	return apiKey, nil
}
