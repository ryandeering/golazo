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
		ID     string `json:"id"`
		Round  string `json:"round"`
		Home   team   `json:"home"`
		Away   team   `json:"away"`
		League league `json:"league"`
	} `json:"general"`
	Content struct {
		MatchFacts struct {
			Events []fotmobEventDetail `json:"events"`
		} `json:"matchFacts"`
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
	// Extract match info from general and header
	baseMatch := fotmobMatch{
		ID:     m.General.ID,
		Round:  m.General.Round,
		Home:   m.General.Home,
		Away:   m.General.Away,
		Status: m.Header.Status,
		League: m.General.League,
	}.toAPIMatch()

	details := &api.MatchDetails{
		Match:  baseMatch,
		Events: make([]api.MatchEvent, 0),
	}

	// Convert events from content.matchFacts.events
	events := make([]api.MatchEvent, 0, len(m.Content.MatchFacts.Events))
	for _, e := range m.Content.MatchFacts.Events {
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
		homeIDInt := parseInt(m.General.Home.ID)
		awayIDInt := parseInt(m.General.Away.ID)
		if e.IsHome {
			event.Team = api.Team{
				ID:        homeIDInt,
				Name:      m.General.Home.Name,
				ShortName: m.General.Home.ShortName,
			}
		} else {
			event.Team = api.Team{
				ID:        awayIDInt,
				Name:      m.General.Away.Name,
				ShortName: m.General.Away.ShortName,
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
