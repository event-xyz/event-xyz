package models

type Events struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Max_Participants int    `json:"max_participants"`
	Event_Date       string `json:"event_date"`
	Form_Link        string `json:"form"`
}
