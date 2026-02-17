package models

// SlideDeck represents a collection of slides for a presentation
type SlideDeck struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Slides      []Slide `json:"slides"`
}

// Slide represents a single slide in the deck
type Slide struct {
	Title        string     `json:"title"`
	Layout       string     `json:"layout"` // title_and_body, two_columns, blank, title_only, chart_focus, data_table
	BulletPoints []string   `json:"bullet_points,omitempty"`
	SpeakerNotes string     `json:"speaker_notes,omitempty"`
	ChartID      string     `json:"chart_id,omitempty"` // ID of the dashboard card/chart
	Headers      []string   `json:"headers,omitempty"`  // Column headers for data_table layout
	Rows         [][]string `json:"rows,omitempty"`     // Data rows for data_table layout
}
