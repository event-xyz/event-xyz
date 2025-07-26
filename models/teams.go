package models

import "time"

type Team struct {
	ID              string    `json:"id"`
	EventID         int       `json:"event_id"`
	Name            string    `json:"name"`
	LeadName        string    `json:"lead_name"`
	LeadEmail       string    `json:"lead_email"`
	LeadPhone       string    `json:"lead_phone"`
	NumParticipants int       `json:"num_participants"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
