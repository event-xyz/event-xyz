package models

// Organiser represents an organiser document in Couchbase.
type DBAuthorisedUsers struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
