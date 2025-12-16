package fotmob

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// fotmobMatch represents a match in FotMob's API format
// Note: FotMob uses string IDs in JSON, but we convert them to ints
type fotmobMatch struct {
	ID     string `json:"id"` // FotMob returns string IDs
	Round  string `json:"round"`
	Home   team   `json:"home"`
	Away   team   `json:"away"`
	Status status `json:"status"`
	League league `json:"league"`
}

type team struct {
	ID        string `json:"id"` // FotMob returns string IDs
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type league struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
}

type status struct {
	UTCTime   string    `json:"utcTime"`   // Can be null/empty
	Started   *bool     `json:"started"`   // Can be null
	Finished  *bool     `json:"finished"`  // Can be null
	Cancelled *bool     `json:"cancelled"` // Can be null
	LiveTime  *liveTime `json:"liveTime,omitempty"`
	Score     *score    `json:"score,omitempty"`
}

type liveTime struct {
	Short string `json:"short"`
}

type score struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// toAPIMatch converts a fotmobMatch to api.Match
func (m fotmobMatch) toAPIMatch() api.Match {
	// Convert string IDs to ints
	matchID := parseInt(m.ID)
	homeID := parseInt(m.Home.ID)
	awayID := parseInt(m.Away.ID)

	match := api.Match{
		ID: matchID,
		League: api.League{
			ID:          m.League.ID,
			Name:        m.League.Name,
			Country:     m.League.Country,
			CountryCode: m.League.CountryCode,
		},
		HomeTeam: api.Team{
			ID:        homeID,
			Name:      m.Home.Name,
			ShortName: m.Home.ShortName,
		},
		AwayTeam: api.Team{
			ID:        awayID,
			Name:      m.Away.Name,
			ShortName: m.Away.ShortName,
		},
		Round: m.Round,
	}

	// Parse match time - FotMob uses .000Z format sometimes
	if m.Status.UTCTime != "" {
		var t time.Time
		var err error
		t, err = time.Parse(time.RFC3339, m.Status.UTCTime)
		if err != nil {
			// Try alternative format with milliseconds
			t, err = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
		}
		if err == nil {
			match.MatchTime = &t
		}
	}

	// Determine status - handle null boolean values
	if m.Status.Cancelled != nil && *m.Status.Cancelled {
		match.Status = api.MatchStatusCancelled
	} else if m.Status.Finished != nil && *m.Status.Finished {
		match.Status = api.MatchStatusFinished
	} else if m.Status.Started != nil && *m.Status.Started {
		match.Status = api.MatchStatusLive
		if m.Status.LiveTime != nil {
			match.LiveTime = &m.Status.LiveTime.Short
		}
	} else {
		match.Status = api.MatchStatusNotStarted
	}

	// Set scores if available
	if m.Status.Score != nil {
		match.HomeScore = &m.Status.Score.Home
		match.AwayScore = &m.Status.Score.Away
	}

	return match
}

// fotmobMatchDetails represents detailed match information from FotMob
// Note: FotMob API returns a nested structure with content.matchFacts containing events
type fotmobMatchDetails struct {
	Header struct {
		Teams []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Score int    `json:"score"`
		} `json:"teams"`
		Status status `json:"status"`
	} `json:"header"`
	General struct {
		MatchID  string `json:"matchId"`
		Round    string `json:"matchRound"`
		HomeTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"homeTeam"`
		AwayTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"awayTeam"`
		LeagueID   int    `json:"leagueId"`
		LeagueName string `json:"leagueName"`
	} `json:"general"`
	Content struct {
		MatchFacts struct {
			Events struct {
				Events []fotmobEventDetail `json:"events"`
			} `json:"events"`
			InfoBox struct {
				Stadium struct {
					Name string `json:"name"`
				} `json:"Stadium,omitempty"`
			} `json:"infoBox,omitempty"`
		} `json:"matchFacts"`
		Stats struct {
			Periods struct {
				All struct {
					Stats []struct {
						Title string `json:"title"`
						Stats []struct {
							Key   string `json:"key"`
							Value string `json:"value"`
						} `json:"stats"`
					} `json:"stats"`
				} `json:"all,omitempty"`
			} `json:"periods,omitempty"`
		} `json:"stats,omitempty"`
	} `json:"content"`
}

// fotmobEventDetail represents a single event detail from FotMob
type fotmobEventDetail struct {
	Time    int    `json:"time"`
	TimeStr int    `json:"timeStr"`
	Type    string `json:"type"`
	EventID int    `json:"eventId"`
	IsHome  bool   `json:"isHome"`
	Player  *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"player,omitempty"`
	PlayerID  *int   `json:"playerId,omitempty"`
	NameStr   string `json:"nameStr,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	FullName  string `json:"fullName,omitempty"`
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
	NewScore  []int  `json:"newScore,omitempty"`
	OwnGoal   *bool  `json:"ownGoal,omitempty"`
	IsPenalty *bool  `json:"isPenalty,omitempty"`
	Card      string `json:"card,omitempty"` // "Yellow" or "Red"
	Swap      []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"swap,omitempty"` // For substitutions
	AssistStr      string `json:"assistStr,omitempty"`
	AssistInput    string `json:"assistInput,omitempty"`
	AssistPlayerID *int   `json:"assistPlayerId,omitempty"`
}

// toAPIMatchDetails converts fotmobMatchDetails to api.MatchDetails
func (m fotmobMatchDetails) toAPIMatchDetails() *api.MatchDetails {
	// Parse match ID from string
	matchID := parseInt(m.General.MatchID)

	// Determine match status from header
	var status api.MatchStatus
	var liveTime *string
	if m.Header.Status.Cancelled != nil && *m.Header.Status.Cancelled {
		status = api.MatchStatusCancelled
	} else if m.Header.Status.Finished != nil && *m.Header.Status.Finished {
		status = api.MatchStatusFinished
	} else if m.Header.Status.Started != nil && *m.Header.Status.Started {
		status = api.MatchStatusLive
		if m.Header.Status.LiveTime != nil {
			liveTime = &m.Header.Status.LiveTime.Short
		}
	} else {
		status = api.MatchStatusNotStarted
	}

	// Parse match time
	var matchTime *time.Time
	if m.Header.Status.UTCTime != "" {
		t, err := time.Parse(time.RFC3339, m.Header.Status.UTCTime)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05.000Z", m.Header.Status.UTCTime)
		}
		if err == nil {
			matchTime = &t
		}
	}

	// Build the base match
	baseMatch := api.Match{
		ID: matchID,
		League: api.League{
			ID:   m.General.LeagueID,
			Name: m.General.LeagueName,
		},
		HomeTeam: api.Team{
			ID:        m.General.HomeTeam.ID,
			Name:      m.General.HomeTeam.Name,
			ShortName: m.General.HomeTeam.Name, // Use full name as short name if not available
		},
		AwayTeam: api.Team{
			ID:        m.General.AwayTeam.ID,
			Name:      m.General.AwayTeam.Name,
			ShortName: m.General.AwayTeam.Name, // Use full name as short name if not available
		},
		Status:    status,
		LiveTime:  liveTime,
		MatchTime: matchTime,
		Round:     m.General.Round,
	}

	details := &api.MatchDetails{
		Match:  baseMatch,
		Events: make([]api.MatchEvent, 0),
	}

	// Populate scores from header.Teams
	if len(m.Header.Teams) >= 2 {
		homeScore := m.Header.Teams[0].Score
		awayScore := m.Header.Teams[1].Score
		details.Match.HomeScore = &homeScore
		details.Match.AwayScore = &awayScore

		// Determine winner for finished matches
		if status == api.MatchStatusFinished {
			if homeScore > awayScore {
				winner := "home"
				details.Winner = &winner
			} else if awayScore > homeScore {
				winner := "away"
				details.Winner = &winner
			}
		}
	}

	// Populate venue from infoBox
	if m.Content.MatchFacts.InfoBox.Stadium.Name != "" {
		details.Venue = m.Content.MatchFacts.InfoBox.Stadium.Name
	}

	// Extract half-time score from events (look for "Half" event type)
	// Also set match duration (default to 90, but can be 120 for extra time)
	details.MatchDuration = 90
	for _, e := range m.Content.MatchFacts.Events.Events {
		if e.Type == "Half" && (e.HomeScore > 0 || e.AwayScore > 0) {
			// Found half-time score
			if details.HalfTimeScore == nil {
				details.HalfTimeScore = &struct {
					Home *int `json:"home,omitempty"`
					Away *int `json:"away,omitempty"`
				}{}
			}
			htHome := e.HomeScore
			htAway := e.AwayScore
			details.HalfTimeScore.Home = &htHome
			details.HalfTimeScore.Away = &htAway
		}
		// Check for extra time indicators (events after 90 minutes)
		if e.Time > 90 {
			details.ExtraTime = true
			details.MatchDuration = 120
		}
	}

	// Populate match events (already being done below, but ensure they're added)
	// Events are converted from content.matchFacts.events

	// Convert events from content.matchFacts.events
	events := make([]api.MatchEvent, 0, len(m.Content.MatchFacts.Events.Events))
	for _, e := range m.Content.MatchFacts.Events.Events {
		// Skip non-event types like "Half"
		if e.Type == "Half" {
			continue
		}

		// Normalize event type to lowercase for consistent matching
		eventType := strings.ToLower(e.Type)

		event := api.MatchEvent{
			ID:        e.EventID,
			Minute:    e.Time,
			Type:      eventType,
			Timestamp: time.Now(),
		}

		// Extract player name
		playerName := ""
		if e.Player != nil && e.Player.Name != "" {
			playerName = e.Player.Name
		} else if e.FullName != "" {
			playerName = e.FullName
		} else if e.NameStr != "" {
			playerName = e.NameStr
		}
		if playerName != "" {
			event.Player = &playerName
		}

		// Extract assist
		if e.AssistInput != "" {
			event.Assist = &e.AssistInput
		}

		// Extract event type details
		eventTypeDetail := ""
		if e.Type == "Card" && e.Card != "" {
			eventTypeDetail = strings.ToLower(e.Card)
		} else if e.Type == "Substitution" && len(e.Swap) >= 2 {
			// Substitution: swap[0] is out, swap[1] is in
			eventTypeDetail = "out"
		}
		if eventTypeDetail != "" {
			event.EventType = &eventTypeDetail
		}

		// Set team based on isHome flag
		homeIDInt := m.General.HomeTeam.ID
		awayIDInt := m.General.AwayTeam.ID
		if e.IsHome {
			event.Team = api.Team{
				ID:        homeIDInt,
				Name:      m.General.HomeTeam.Name,
				ShortName: m.General.HomeTeam.Name,
			}
		} else {
			event.Team = api.Team{
				ID:        awayIDInt,
				Name:      m.General.AwayTeam.Name,
				ShortName: m.General.AwayTeam.Name,
			}
		}

		events = append(events, event)
	}

	// Sort events by minute (chronological order)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Minute < events[j].Minute
	})

	details.Events = events
	return details
}

// fotmobTableRow represents a single row in the league table from FotMob
type fotmobTableRow struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ShortName      string `json:"shortName"`
	Rank           int    `json:"rank"`
	Played         int    `json:"played"`
	Wins           int    `json:"wins"`
	Draws          int    `json:"draws"`
	Losses         int    `json:"losses"`
	GoalsFor       int    `json:"goalsFor"`
	GoalsAgainst   int    `json:"goalsAgainst"`
	GoalDifference int    `json:"goalDifference"`
	Points         int    `json:"points"`
}

// toAPITableEntry converts fotmobTableRow to api.LeagueTableEntry
func (r fotmobTableRow) toAPITableEntry() api.LeagueTableEntry {
	return api.LeagueTableEntry{
		Position: r.Rank,
		Team: api.Team{
			ID:        r.ID,
			Name:      r.Name,
			ShortName: r.ShortName,
		},
		Played:         r.Played,
		Won:            r.Wins,
		Drawn:          r.Draws,
		Lost:           r.Losses,
		GoalsFor:       r.GoalsFor,
		GoalsAgainst:   r.GoalsAgainst,
		GoalDifference: r.GoalDifference,
		Points:         r.Points,
	}
}

// Helper function to parse time from various formats
func parseTime(timeStr string) *time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return &t
		}
	}
	return nil
}

// Helper function to parse int from string
// Returns 0 if parsing fails (for required fields)
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}
