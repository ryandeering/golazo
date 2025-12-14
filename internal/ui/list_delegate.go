package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// NewMatchListDelegate creates a custom list delegate for match items.
func NewMatchListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Customize styles - use highlight color for selected items
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(highlightColor).
		Bold(true).
		Padding(0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(highlightColor).
		Padding(0, 1)

	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(textColor).
		Padding(0, 1)
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(dimColor).
		Padding(0, 1)

	return d
}

// MatchListStyles returns styles for the list component.
func MatchListStyles() lipgloss.Style {
	return panelStyle
}
