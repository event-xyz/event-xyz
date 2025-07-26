package models

type Participant struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"` // Reference to Team ID
	Role        string `json:"role"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	College     string `json:"college"`
	SRN         string `json:"srn"`
	Branch      string `json:"branch"`
	DayScholar  bool   `json:"day_scholar"`      // Bool value (true/false)
	Hostel      string `json:"hostel,omitempty"` // Nullable field
	Shortlisted bool   `json:"shortlisted"`      // Bool value (true/false)
	QRString    string `json:"qr_string"`
}
