package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type EventDB struct {
	Db *gorm.DB
}

// Rather proposed custom error types
var (
	ErrDbOpenFailure   = fmt.Errorf("failed to run `OpenDB()`")
	ErrDbMissingRecord = fmt.Errorf("failed to fetch record")
	ErrIncorrectField  = fmt.Errorf("checkpoint missing in db")
)

// Event specific errors
// DB_Participants related errors
var (
	ErrParticipantAbsent = fmt.Errorf("participant never checkedin")
	ErrParticipantLeft   = fmt.Errorf("participant has left the event")
	ErrCheckpointCrossed = fmt.Errorf("participant has already cleared the checkpoint")
)

// Event authorised user specific
var (
	ErrUnauthorised = fmt.Errorf("incoming request isn't from an authorised login")
	ErrNoAccess     = fmt.Errorf("incoming request was authoriesed but has no access to the endpoint")
)

func TryInitializeDB(dbpath string) (*EventDB, error) {
	if dbpath == "" {
		dbpath = "event.db"
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  true,          // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = db.AutoMigrate(&Checkpoints{})
	if err != nil {
		log.Fatalln("Failed to migrate db Checkpoints")
	}

	err = db.AutoMigrate(&Team{})
	if err != nil {
		log.Fatalln("Failed to migrate db Team")
	}

	err = db.AutoMigrate(&Participant{})
	if err != nil {
		log.Fatalln("Failed to migrate db Participant")
	}

	err = db.AutoMigrate(&DBParticipant{})
	if err != nil {
		log.Fatalln("Failed to migrate db DBParticipant")
	}

	err = db.AutoMigrate(&DBAuthorisedUsers{})
	if err != nil {
		log.Fatalln("Failed to migrate db DBAuthoriesedUsers")
	}

	err = db.AutoMigrate(&ClaimsLogs{})
	if err != nil {
		log.Fatalln("Failed to migrate db ClaimsLogs")
	}

	log.Println("[CRUD] InitializeDB and Migraated tables sucessfully")
	return &EventDB{
		Db: db,
	}, nil
}

// Open and return db access struct
func (e *EventDB) openDB() (*gorm.DB, error) {
	if e.Db == nil {
		log.Fatalln("Database not initilised or loaded")
	}
	return e.Db, nil
}

func (e *EventDB) CreateTeam(team *Team) error {
	db, err := e.openDB()
	if err != nil {
		return err
	}

	db.Create(&team)
	log.Printf("Creating team %s with id %d", team.Team, team.ID)

	return nil
}

// Create records for all participants parsed from the csv
func (e *EventDB) CreateParticipants(dbParticipants []DBParticipant) error {
	db, err := e.openDB()
	if err != nil {
		return err
	}

	log.Println("Attempting to write to db")
	// Writing to DB
	db.Create(dbParticipants)

	return nil
}

func CheckpointsWithDefaults() *Checkpoints {
	return &Checkpoints{
		Checkin:   false,
		Checkout:  false,
		Snacks:    false,
		Dinner:    false,
		Breakfast: false,
	}
}

func (e *EventDB) CreateParticipant(
	participant *Participant,
	checkpoints *Checkpoints,
	uuid string,
	claims string,
) error {
	db, err := e.openDB()
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Create(&checkpoints)
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = tx.Create(&participant)
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = tx.Create(&DBParticipant{
			UUID:          uuid,
			ParticipantID: participant.ID,
			CheckpointsID: checkpoints.ID,
		})
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = tx.Create(&ClaimsLogs{
			Jwt:           claims,
			ParticipantID: participant.ID,
		})
		if txErr.Error != nil {
			return txErr.Error
		}

		return nil
	})

	return err
}

func (e *EventDB) CreateVolunteer(volunteer *DBAuthorisedUsers) error {
	db, err := e.openDB()
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Create(&volunteer)
		if txErr.Error != nil {
			return txErr.Error
		}

		return nil
	})

	return nil
}

// Returns (success, error) if success is zero then
// if no rows were affected this means that
// all three (name, phone and email) were wrong.
// send a warning to the frontend by setting success as 0.
func (e *EventDB) CreateVolunteerManual(volunteer *DBAuthorisedUsers) (int, error) {
	// Initially set success as 1
	success := 1

	db, err := e.openDB()
	if err != nil {
		return 0, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("name = ? OR phone = ? OR verified_email = ?",
			volunteer.Name, volunteer.Phone, volunteer.VerifiedEmail).
			Delete(&DBAuthorisedUsers{})

		if result.RowsAffected == 0 {
			success = 0
			log.Println("Warning: No existing volunteer found for deletion (manual entry)")
		}

		txErr := tx.Create(&volunteer)
		if txErr.Error != nil {
			return txErr.Error
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return success, nil
}

func (e *EventDB) CreateAuthorisedUsersDB() error {
	db, err := e.openDB()
	if err != nil {
		return ErrDbOpenFailure
	}

	var dbAuth []DBAuthorisedUsers

	db.Create(&dbAuth)

	return nil
}

// Function to fetch all checkpoints and
// forward to backend
//
// for dynamic checkpoint loading for site
func FetchCheckpoints() []string {
	var checkpoints []string

	return checkpoints
}

func (e *EventDB) VerifyLogin(userDetails DBAuthorisedUsers) (*DBAuthorisedUsers, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}

	log.Println("request: ", userDetails)

	var dbAuthUser DBAuthorisedUsers
	_ = db.First(&dbAuthUser, "sub = ?", userDetails.SUB)

	if dbAuthUser.VerifiedEmail == "" {

		_ = db.First(&dbAuthUser, "verified_email = ?", userDetails.VerifiedEmail)

		switch dbAuthUser.UserRole {
		case "admin":
			{
				log.Println("Hello admin ", userDetails.VerifiedEmail)
			}
		case "organisers":
			{
				log.Println("Hello organiser ", userDetails.VerifiedEmail)
			}
		case "volunteer":
			{
				log.Println("Hello volunteer ", userDetails.VerifiedEmail)
			}
		default:
			return nil, ErrDbMissingRecord
		}

		dbAuthUser.SUB = userDetails.SUB
		db.Save(&dbAuthUser)

		return &dbAuthUser, nil
	}

	log.Println("user exists in dbAuthUser")

	return &dbAuthUser, nil
}

func (e *EventDB) SubAuthentication(sub string, userRole string) (*DBAuthorisedUsers, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}

	var dbAuthUser DBAuthorisedUsers
	_ = db.First(&dbAuthUser, "sub = ?", sub)

	if dbAuthUser.VerifiedEmail == "" {
		// bro doesnt exist
		return nil, ErrDbMissingRecord
	}

	if dbAuthUser.UserRole != userRole {
		return nil, ErrNoAccess
	} else {
		return &dbAuthUser, nil
	}
}

func (e *EventDB) JWTFetchParticipant(jwt string) (*DBParticipant, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	_ = db.First(&dbParticipant, "id = ?", jwt)

	return &dbParticipant, nil
}

func (e *EventDB) FetchParticipant(name string, phone string) (*Participant, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}
	var participant Participant

	err = db.Transaction(func(tx *gorm.DB) error {
		_ = tx.First(&participant, "name = ? and phone = ?", name, phone)
		log.Println(participant)

		if participant.Name == "" {
			return ErrDbMissingRecord
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &participant, err
}

// Update DB with the participant entry checkpoint
//
// Return signature
// - Pointer to participant
// - checkin : true if not checked in
// - error
func (e *EventDB) ParticipantEntry(p_uuid string) (*DBParticipant, bool, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, false, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	flag := true

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Debug().Model(&dbParticipant).Preload("Participant").Preload("Checkpoints").First(&dbParticipant, "uuid = ?", p_uuid)
		if txErr.Error != nil {
			return txErr.Error
		}

		if dbParticipant.Participant.Name == "" {
			flag = false
			return ErrDbMissingRecord
		}

		if dbParticipant.Checkpoints.Checkin && dbParticipant.Checkpoints.Checkout {
			flag = false
			return ErrParticipantLeft
		}

		if !dbParticipant.Checkpoints.Checkin {
			dbParticipant.Checkpoints.Entry_time = time.Now()
			dbParticipant.Checkpoints.Checkin = true
			flag = true
			db.Save(&dbParticipant.Checkpoints)
			return nil
		}

		return nil
	})

	return &dbParticipant, flag, err
}

// Update DB with the participant exit checkpoint
//
// Return signature
//   - Pointer to participant
//   - checkin : true if sucessfulyl checked in and
//     false if alreayd checked in
//   - error
func (e *EventDB) ParticipantExit(p_uuid string) (*DBParticipant, bool, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, false, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	flag := true

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Debug().Model(&dbParticipant).Preload("Participant").Preload("Checkpoints").First(&dbParticipant, "uuid = ?", p_uuid)
		if txErr.Error != nil {
			return txErr.Error
		}

		if dbParticipant.Participant.Name == "" {
			return ErrDbMissingRecord
		}

		if !dbParticipant.Checkpoints.Checkin {
			return ErrParticipantAbsent
		}

		if !dbParticipant.Checkpoints.Checkout {
			dbParticipant.Checkpoints.Checkout = true
			dbParticipant.Checkpoints.Exit_time = time.Now()
			db.Save(&dbParticipant.Checkpoints)
			return nil
		}

		return ErrParticipantLeft
	})

	return &dbParticipant, flag, err
}

func (e *EventDB) ParticipantCheckpoint(p_uuid string, checkpointName string) (*DBParticipant, bool, error) {
	db, err := e.openDB()
	if err != nil {
		return nil, false, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	flag := true

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Debug().Model(&dbParticipant).Preload("Participant").Preload("Checkpoints").First(&dbParticipant, "uuid = ?", p_uuid)
		if txErr.Error != nil {
			return txErr.Error
		}

		if dbParticipant.Participant.Name == "" {
			flag = false
			return ErrDbMissingRecord
		}

		if !dbParticipant.Checkpoints.Checkin {
			flag = false
			return ErrParticipantAbsent
		}

		if dbParticipant.Checkpoints.Checkin && dbParticipant.Checkpoints.Checkout {
			flag = false
			return ErrParticipantLeft
		}

		switch checkpointName {
		case "Breakfast":
			{
				if dbParticipant.Checkpoints.Breakfast {
					return ErrCheckpointCrossed
				}
				dbParticipant.Checkpoints.Breakfast = true
				db.Save(&dbParticipant.Checkpoints)
				flag = true
				return nil
			}
		case "Dinner":
			{
				if dbParticipant.Checkpoints.Dinner {
					return ErrCheckpointCrossed
				}
				dbParticipant.Checkpoints.Dinner = true
				db.Save(&dbParticipant.Checkpoints)
				flag = true
				return nil
			}
		case "Snacks":
			{
				if dbParticipant.Checkpoints.Snacks {
					return ErrCheckpointCrossed
				}
				dbParticipant.Checkpoints.Snacks = true
				db.Save(&dbParticipant.Checkpoints)
				flag = true
				return nil
			}
		default:
			{
				flag = false
				return ErrIncorrectField
			}
		}
	})

	// Participant has already opted for the option
	return &dbParticipant, flag, err
}
