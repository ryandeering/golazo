package constants

// Menu items
const (
	MenuStats       = "Stats"
	MenuLiveMatches = "Live Matches"
)

// Panel titles
const (
	PanelLiveMatches     = "Live Matches"
	PanelFinishedMatches = "Finished Matches"
	PanelMinuteByMinute  = "Minute-by-minute"
	PanelMatchStatistics = "Match Statistics"
	PanelUpdates         = "Updates"
)

// Empty state messages
const (
	EmptyNoLiveMatches     = "No live matches"
	EmptyNoFinishedMatches = "No finished matches"
	EmptySelectMatch       = "Select a match"
	EmptyNoUpdates         = "No updates"
	EmptyNoMatches         = "No matches available"
	EmptyAPIKeyMissing     = "API key not configured\n\nSet FOOTBALL_DATA_API_KEY environment variable:\n  export FOOTBALL_DATA_API_KEY=\"your-key-here\"\n\nGet your free API key at:\n  https://www.api-sports.io/"
)

// Help text
const (
	HelpMainMenu    = "↑/↓: navigate  Enter: select  q: quit"
	HelpMatchesView = "↑/↓: navigate  Esc: back  q: quit"
)

// Status text
const (
	StatusLive            = "LIVE"
	StatusFinished        = "FT"
	StatusNotStarted      = "VS"
	StatusNotStartedShort = "NS"
	StatusFinishedText    = "Finished"
)

// Loading text
const (
	LoadingFetching = "Fetching..."
)

// Stats labels
const (
	LabelStatus = "Status: "
	LabelScore  = "Score: "
	LabelLeague = "League: "
	LabelDate   = "Date: "
	LabelVenue  = "Venue: "
)
