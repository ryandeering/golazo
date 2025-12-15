package fotmob

import (
	"context"
	"fmt"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// FinishedMatchesByDateRange retrieves finished matches within a date range.
// This is used for the stats view to show completed matches.
// Queries each date individually and aggregates results, as FotMob doesn't support date range queries.
func (c *Client) FinishedMatchesByDateRange(ctx context.Context, dateFrom, dateTo time.Time) ([]api.Match, error) {
	var allMatches []api.Match

	// Normalize dates to UTC to avoid timezone issues
	currentDate := dateFrom.UTC()
	dateToUTC := dateTo.UTC()
	var lastErr error
	successCount := 0
	totalDates := 0

	for !currentDate.After(dateToUTC) {
		totalDates++
		dateStr := currentDate.Format("2006-01-02")

		// Query matches for this date using MatchesByDate
		matches, err := c.MatchesByDate(ctx, currentDate)
		if err != nil {
			lastErr = fmt.Errorf("fetch matches for date %s: %w", dateStr, err)
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		successCount++
		// Filter for finished matches only
		for _, match := range matches {
			if match.Status == api.MatchStatusFinished {
				allMatches = append(allMatches, match)
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Return error if all dates failed, but allow partial results
	if successCount == 0 && totalDates > 0 {
		return allMatches, fmt.Errorf("failed to fetch matches for any date in range %s to %s: %w", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"), lastErr)
	}

	return allMatches, nil
}

// RecentFinishedMatches retrieves finished matches from the last N days.
// Queries from today going back N-1 days (inclusive).
// Uses UTC timezone to ensure consistent date calculations.
func (c *Client) RecentFinishedMatches(ctx context.Context, days int) ([]api.Match, error) {
	if days <= 0 {
		days = 1 // Default to 1 day if invalid
	}
	// Use UTC to avoid timezone issues
	today := time.Now().UTC()
	dateFrom := today.AddDate(0, 0, -(days - 1)) // Go back (days-1) days to include today
	return c.FinishedMatchesByDateRange(ctx, dateFrom, today)
}

// UpcomingMatches retrieves upcoming (not started) matches for today.
// This is used for the stats view when 1-day period is selected.
// Only fetches matches that haven't started yet.
func (c *Client) UpcomingMatches(ctx context.Context) ([]api.Match, error) {
	today := time.Now().UTC()
	dateStr := today.Format("2006-01-02")

	// Query all matches for today using MatchesByDate
	matches, err := c.MatchesByDate(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("fetch matches for date %s: %w", dateStr, err)
	}

	// Filter for upcoming matches (not started and not finished)
	var upcomingMatches []api.Match
	for _, match := range matches {
		// Include matches that are not started and not finished
		if match.Status == api.MatchStatusNotStarted {
			upcomingMatches = append(upcomingMatches, match)
		}
	}

	return upcomingMatches, nil
}
