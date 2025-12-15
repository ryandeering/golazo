package ui

// Truncate truncates text to fit the specified width, appending "..." if truncated.
func Truncate(text string, width int) string {
	if len(text) <= width {
		return text
	}
	return text[:width-3] + "..."
}
