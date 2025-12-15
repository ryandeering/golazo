package data

import (
	"encoding/json"
	"os"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockMatchesData contains multiple matches for testing.
// Uses api.Match directly to ensure single source of truth.
type MockMatchesData struct {
	Matches []api.Match `json:"matches"`
}

// MockMatches returns multiple mock matches for testing.
func MockMatches() ([]api.Match, error) {
	// Try to load from config directory first
	configPath, err := MockDataPath()
	if err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			var mockData MockMatchesData
			if err := json.Unmarshal(data, &mockData); err == nil {
				return mockData.Matches, nil
			}
		}
	}

	// Fallback to embedded mock data
	return getDefaultMockMatches(), nil
}

// getDefaultMockMatches returns default mock matches.
func getDefaultMockMatches() []api.Match {
	matches := []api.Match{
		{
			ID: 1,
			League: api.League{
				ID:          1,
				Name:        "Premier League",
				Country:     "England",
				CountryCode: "GB",
			},
			HomeTeam: api.Team{
				ID:        1,
				Name:      "Manchester United",
				ShortName: "Man Utd",
			},
			AwayTeam: api.Team{
				ID:        2,
				Name:      "Liverpool",
				ShortName: "Liverpool",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(2),
			AwayScore: intPtr(1),
			LiveTime:  stringPtr("67'"),
			Round:     "Matchday 15",
		},
		{
			ID: 2,
			League: api.League{
				ID:          2,
				Name:        "La Liga",
				Country:     "Spain",
				CountryCode: "ES",
			},
			HomeTeam: api.Team{
				ID:        3,
				Name:      "Real Madrid",
				ShortName: "Real Madrid",
			},
			AwayTeam: api.Team{
				ID:        4,
				Name:      "Barcelona",
				ShortName: "Barcelona",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(1),
			AwayScore: intPtr(1),
			LiveTime:  stringPtr("23'"),
			Round:     "Matchday 12",
		},
		{
			ID: 3,
			League: api.League{
				ID:          3,
				Name:        "Serie A",
				Country:     "Italy",
				CountryCode: "IT",
			},
			HomeTeam: api.Team{
				ID:        5,
				Name:      "AC Milan",
				ShortName: "AC Milan",
			},
			AwayTeam: api.Team{
				ID:        6,
				Name:      "Inter Milan",
				ShortName: "Inter",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(3),
			AwayScore: intPtr(2),
			LiveTime:  stringPtr("FT"),
			Round:     "Matchday 10",
		},
		{
			ID: 4,
			League: api.League{
				ID:          1,
				Name:        "Premier League",
				Country:     "England",
				CountryCode: "GB",
			},
			HomeTeam: api.Team{
				ID:        7,
				Name:      "Arsenal",
				ShortName: "Arsenal",
			},
			AwayTeam: api.Team{
				ID:        8,
				Name:      "Chelsea",
				ShortName: "Chelsea",
			},
			Status: api.MatchStatusNotStarted,
			Round:  "Matchday 15",
		},
	}

	// Save to config directory for persistence
	if configPath, err := MockDataPath(); err == nil {
		mockData := MockMatchesData{Matches: matches}
		if data, err := json.MarshalIndent(mockData, "", "  "); err == nil {
			os.WriteFile(configPath, data, 0644)
		}
	}

	return matches
}

func intPtr(i int) *int {
	return &i
}
