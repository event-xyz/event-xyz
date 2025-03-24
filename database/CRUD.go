package database

import (
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DbGlobal  *gorm.DB
	errGlobal error
)

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
)

// Event authorised user specific
var (
	ErrUnauthorised = fmt.Errorf("incoming request isn't from an authorised login")
	ErrNoAccess     = fmt.Errorf("incoming request was authoriesed but has no access to the endpoint")
)

func InitializeDB() error {
	if DbGlobal == nil {
		dbPath := os.Getenv("DBPATH")
		if dbPath == "" {
			dbPath = "event.db"
		}
		DbGlobal, errGlobal = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if errGlobal != nil {
			log.Println(errGlobal)
			return errGlobal
		}

		err := DbGlobal.AutoMigrate(&Checkpoints{})
		if err != nil {
			log.Fatalln("Failed to migrate db Checkpoints")
		}

		err = DbGlobal.AutoMigrate(&Team{})
		if err != nil {
			log.Fatalln("Failed to migrate db Team")
		}

		err = DbGlobal.AutoMigrate(&Participant{})
		if err != nil {
			log.Fatalln("Failed to migrate db Participant")
		}

		err = DbGlobal.AutoMigrate(&DBParticipant{})
		if err != nil {
			log.Fatalln("Failed to migrate db DBParticipant")
		}

		err = DbGlobal.AutoMigrate(&DBAuthoriesedUsers{})
		if err != nil {
			log.Fatalln("Failed to migrate db DBAuthoriesedUsers")
		}

		err = DbGlobal.AutoMigrate(&ClaimsLogs{})
		if err != nil {
			log.Fatalln("Failed to migrate db ClaimsLogs")
		}
	}
	log.Println("[CRUD] InitializeDB and Migraated tables sucessfully")
	return nil
}

// Open and return db access struct
func openDB() (*gorm.DB, error) {
	if DbGlobal == nil {
		dbPath := os.Getenv("DBPATH")
		if dbPath == "" {
			dbPath = "event.db"
		}
		DbGlobal, errGlobal = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if errGlobal != nil {
			log.Println(errGlobal)
			return nil, errGlobal
		}

		err := DbGlobal.AutoMigrate(&Checkpoints{})
		if err != nil {
			log.Fatalln("Failed to migrate db Checkpoints")
		}

		err = DbGlobal.AutoMigrate(&Team{})
		if err != nil {
			log.Fatalln("Failed to migrate db Team")
		}

		err = DbGlobal.AutoMigrate(&Participant{})
		if err != nil {
			log.Fatalln("Failed to migrate db Participant")
		}

		err = DbGlobal.AutoMigrate(&DBParticipant{})
		if err != nil {
			log.Fatalln("Failed to migrate db DBParticipant")
		}

		err = DbGlobal.AutoMigrate(&DBAuthoriesedUsers{})
		if err != nil {
			log.Fatalln("Failed to migrate db DBAuthoriesedUsers")
		}

		err = DbGlobal.AutoMigrate(&ClaimsLogs{})
		if err != nil {
			log.Fatalln("Failed to migrate db ClaimsLogs")
		}
	}
	return DbGlobal, nil
}

func CreateTeam(team *Team) error {
	db, err := openDB()
	if err != nil {
		return err
	}

	db.Create(&team)
	log.Printf("Creating team %s with id %d", team.Team, team.ID)

	return nil
}

// Create records for all participants parsed from the csv
func CreateParticipants(dbParticipants []DBParticipant) error {
	db, err := openDB()
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

func CreateParticipant(
	participant *Participant,
	checkpoints *Checkpoints,
	uuid string,
	claims string,
) error {
	db, err := openDB()
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

func CreateAuthorisedUsersDB() error {
	db, err := openDB()
	if err != nil {
		return ErrDbOpenFailure
	}

	var dbAuth []DBAuthoriesedUsers

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

func VerifyLogin(userDetails DBAuthoriesedUsers) (*DBAuthoriesedUsers, error) {
	db, err := openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}

	log.Println("request: ", userDetails)

	var dbAuthUser DBAuthoriesedUsers
	_ = db.First(&dbAuthUser, "sub = ?", userDetails.SUB)

	if dbAuthUser.VerifiedEmail == "" {

		log.Println("Missing from records")

		// need admin side approval for
		// organisers and volunteers
		admins := []string{
			"adityahegde.clg@gmail.com",
			"adheshathrey2004@gmail.com",
		}

		organisers := []string{
			"anirudh.sudhir1@gmail.com",
		}

		volunteers := []string{
			"adimhegde@gmail.com",
			"naysha.k0708@gmail.com",
			"devesh6742@gmail.com",
			"omshivshankar21@gmail.com",
			"vickspatil1404@gmail.com",
			"kunalkishoremaverick@gmail.com",
			"kavyaprakashscei@gmail.com",
			"nehanshetty2003@gmail.com",
			"b.himank101@gmail.com",
			"moulikmachaiah724@gmail.com",
			"manum262sagara@gmail.com",
			"disha14072003@gmail.com",
			"ananya975.p@gmail.com",
			"prathamshetty0826@gmail.com",
			"ruthu.hm03@gmail.com",
			"shashanknadigm03@gmail.com",
			"keerthanaumesh161@gmail.com",
			"shreyalizbethrobin@gmail.com",
			"eshwarra5@gmail.com",
			"jiteshnayak2004@gmail.com",
			"sarkarsoham73@gmail.com",
			"santoshrajpurohit89@gmail.com",
			"prawnee99@gmail.com",
			"shubhammookim@gmail.com",
			"kushagraagarwal2003@gmail.com",
			"roshinlinson67281@gmail.com",
		}

		if slices.Contains(admins, userDetails.VerifiedEmail) {
			userDetails.UserRole = "admin"
			log.Println("Hello admin")
		} else if slices.Contains(organisers, userDetails.VerifiedEmail) {
			userDetails.UserRole = "organiser"
			log.Println("Hello organiser")
		} else if slices.Contains(volunteers, userDetails.VerifiedEmail) {
			userDetails.UserRole = "volunteer"
		} else {
			return nil, ErrDbMissingRecord
		}

		db.Create(userDetails)
		return &userDetails, nil
	}

	return &dbAuthUser, nil
}

func SubAuthentication(sub string, userRole string) (*DBAuthoriesedUsers, error) {
	db, err := openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}

	var dbAuthUser DBAuthoriesedUsers
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

func JWTFetchParticipant(jwt string) (*DBParticipant, error) {
	db, err := openDB()
	if err != nil {
		return nil, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	_ = db.First(&dbParticipant, "id = ?", jwt)

	return &dbParticipant, nil
}

func FetchParticipant(name string, phone string) (*Participant, error) {
	db, err := openDB()
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
func ParticipantEntry(p_uuid string) (*DBParticipant, *Participant, *Checkpoints, bool, error) {
	db, err := openDB()
	if err != nil {
		return nil, nil, nil, false, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	var participant Participant
	var checkpoint Checkpoints
	flag := true

	err = db.Transaction(func(tx *gorm.DB) error {
		_ = db.First(&dbParticipant, "uuid = ?", p_uuid)
		log.Println(dbParticipant)

		_ = db.First(&participant, "id = ?", dbParticipant.ParticipantID)
		log.Println(participant)

		_ = db.First(&checkpoint, "id = ?", dbParticipant.CheckpointsID)
		log.Println(checkpoint)

		if participant.Name == "" {
			flag = false
			return ErrDbMissingRecord
		}

		if checkpoint.Checkin && checkpoint.Checkout {
			flag = false
			return ErrParticipantLeft
		}

		if !checkpoint.Checkin {
			checkpoint.Entry_time = time.Now()
			checkpoint.Checkin = true
			flag = true
			db.Save(&checkpoint)
			return nil
		}

		return nil
	})

	return &dbParticipant, &participant, &checkpoint, flag, err
}

// Update DB with the participant exit checkpoint
//
// Return signature
//   - Pointer to participant
//   - checkin : true if sucessfulyl checked in and
//     false if alreayd checked in
//   - error
func ParticipantExit(p_uuid string) (*DBParticipant, *Participant, *Checkpoints, bool, error) {
	db, err := openDB()
	if err != nil {
		return nil, nil, nil, false, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	var participant Participant
	var checkpoint Checkpoints
	flag := true

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.First(&dbParticipant, "uuid = ?", p_uuid)
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = db.First(&participant, "id = ? ", dbParticipant.ParticipantID)
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = db.First(&checkpoint, "id = ? ", dbParticipant.CheckpointsID)
		if txErr.Error != nil {
			return txErr.Error
		}

		if participant.Name == "" {
			return ErrDbMissingRecord
		}

		if !checkpoint.Checkin {
			return ErrParticipantAbsent
		}

		if !checkpoint.Checkout {
			checkpoint.Checkout = true
			checkpoint.Exit_time = time.Now()
			db.Save(checkpoint)
			return nil
		}

		return ErrParticipantLeft
	})

	return &dbParticipant, &participant, &checkpoint, flag, err
}

func ParticipantCheckpoint(p_uuid string, checkpointName string) (*DBParticipant, *Participant, *Checkpoints, bool, error) {
	db, err := openDB()
	if err != nil {
		return nil, nil, nil, false, ErrDbOpenFailure
	}

	var dbParticipant DBParticipant
	var participant Participant
	var checkpoint Checkpoints
	flag := true

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.First(&dbParticipant, "uuid = ?", p_uuid)
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = db.First(&participant, "id = ?", dbParticipant.ParticipantID)
		if txErr.Error != nil {
			return txErr.Error
		}

		txErr = db.First(&checkpoint, "id = ?", dbParticipant.CheckpointsID)
		if txErr.Error != nil {
			return txErr.Error
		}

		if participant.Name == "" {
			flag = false
			return ErrDbMissingRecord
		}

		if !checkpoint.Checkin {
			flag = false
			return ErrParticipantAbsent
		}

		if checkpoint.Checkin && checkpoint.Checkout {
			flag = false
			return ErrParticipantLeft
		}

		// FIX: Refractor
		switch checkpointName {
		case "Breakfast":
			{
				if checkpoint.Breakfast {
					break
				}
				checkpoint.Breakfast = true
				db.Save(&checkpoint)
				flag = true
				return nil
			}
		case "Dinner":
			{
				if checkpoint.Dinner {
					break
				}
				checkpoint.Dinner = true
				db.Save(&checkpoint)
				flag = true
				return nil
			}
		case "Snacks":
			{
				if checkpoint.Snacks {
					break
				}
				checkpoint.Snacks = true
				db.Save(&checkpoint)
				flag = true
				return nil
			}
		default:
			{
				flag = false
				return ErrIncorrectField
			}
		}

		return nil
	})

	// Check if participant is in the db

	// Participant has already opted for the option
	return &dbParticipant, &participant, &checkpoint, flag, err
}
