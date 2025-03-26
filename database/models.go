package database

import "time"

// This struct is another table to hold team data
type Team struct {
	// TODO: update by using a UUID
	ID    int64  `gorm:"primaryKey;not null;autoIncrement"`
	Team  string `gorm:"not null"`
	Email string `gorm:"not null"`

	Participant []Participant
}

const (
	AccessRoleAdmin     = iota
	AccessRoleOrganizer = iota
	AccessRoleVolunteer = iota
)

// This struct represents the verified organizers and
// volunteers, holding the UUID (sub part of payload of
// google OAuth), VerifiedEmail gmails and user roles
// allowed roles:
//   - admain
//   - organizers
//   - volunteers
type DBAuthoriesedUsers struct {
	ID            int64 `gorm:"primaryKey;not null;autoIncrement"`
	SUB           string
	VerifiedEmail string
	UserRole      string
	Name          string
	Phone         int64

	// TOOD:
	// event name they belong to
}

// This struct holds all of the unified participant data stored in the DB
type DBParticipant struct {
	// ID is the generated JWT auth token
	// TODO: Change to a UUID instead of JWT
	ID   int64  `gorm:"primaryKey;not null;autoIncrement"`
	UUID string `gorm:"unique"`

	// foreign keys
	ParticipantID int64
	Participant   Participant
	CheckpointsID int64
	Checkpoints   Checkpoints
}

// This struct stores all fields parsed from the csv
type Participant struct {
	// TODO: Seperate to a separate
	// team table
	ID        int64  `gorm:"primaryKey;not null;autoIncrement"`
	Name      string `json:"name"`
	Phone     int64  `json:"phone"`
	Branch    string `json:"branch"`
	PesHostel string `json:"pesHostel"`

	// foreign keys
	TeamID int64
	Team   Team
}

// This struct stores all event checkpoints
type Checkpoints struct {
	// Other Participant Parameters
	ID         int64     `gorm:"primaryKey;not null;autoIncrement"`
	Entry_time time.Time `json:"entry_time"`
	Checkin    bool      `json:"checkin"`
	Checkout   bool      `json:"checkout"`
	Exit_time  time.Time `json:"exit_time"`

	// TODO: Check with organizers regarding
	// the checkpoints of events
	Snacks    bool `json:"snacks"`
	Dinner    bool `json:"dinner"`
	Breakfast bool `json:"breakfast"`
}

type ClaimsLogs struct {
	ID            int64  `gorm:"primaryKey;not null;autoIncrement"`
	Jwt           string `gorm:"unique"`
	ParticipantID int64
	Participant   Participant
}
