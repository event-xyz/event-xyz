package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/homebrew-ec-foss/eventloop/database"
	"gorm.io/gorm"
)

type ParticipantInfo struct {
	database.Participant
	database.Team
}

type ParticipantCheckpointInfo struct {
	database.Participant
	database.Checkpoints
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func (a *App) handleFormData(ctx *gin.Context, userRole string) {
	log.Println("Middleware auth for multipart/form-data")

	authJson, err := ctx.FormFile("sub")
	if err != nil {
		ctx.String(http.StatusBadRequest, "Error: No auth json")
		ctx.Abort()
		return
	}

	authJsonFile, err := authJson.Open()
	if err != nil {
		ctx.String(http.StatusBadRequest, "Error: Failed to open auth json content")
		ctx.Abort()
		return
	}
	defer authJsonFile.Close()

	reader := bufio.NewReader(authJsonFile)
	content := bytes.Buffer{}
	_, err = io.Copy(&content, reader)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error: Failed to read auth file")
		ctx.Abort()
		return
	}

	var subBlobJson map[string]any
	err = json.Unmarshal(content.Bytes(), &subBlobJson)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error: Failed to parse blob to json")
		ctx.Abort()
		return
	}

	_, err = a.Store.SubAuthentication(subBlobJson["sub"].(string), "admin")

	switch err {
	case database.ErrDbOpenFailure:
		{
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Backend server failed to perform operations. Contact administrator"})
			ctx.Abort()
			return
		}
	case database.ErrDbMissingRecord:
		{
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Missing auth records for incoming user"})
			ctx.Abort()
			return
		}
	case database.ErrNoAccess:
		{
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Incoming request was authoriesed but has no access to endpoint"})
			ctx.Abort()
			return
		}
	}
	log.Println("Authorised request")
}

func (a *App) handleApplicationJson(ctx *gin.Context, userRole string) {
	log.Println("Middleware auth for application/json")

	body, err := io.ReadAll(ctx.Request.Body)
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		ctx.Abort()
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
		ctx.Abort()
		return
	}

	log.Println(data)

	user_sub := data["sub"].(string)

	_, err = a.Store.SubAuthentication(user_sub, userRole)

	switch err {
	case database.ErrDbOpenFailure:
		{
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Backend server failed to perform operations. Contact administrator"})
			ctx.Abort()
			return
		}
	case database.ErrDbMissingRecord:
		{
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Missing auth records for incoming user"})
			ctx.Abort()
			return
		}
	case database.ErrNoAccess:
		{
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Incoming request was authoriesed but has no access to endpoint"})
			ctx.Abort()
			return
		}
	}

	// querry db and check if they have the grant
	log.Println("Authorised request")
}

func (a *App) AuthenticationMiddleware(userRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		contType := strings.Split(ctx.Request.Header.Get("Content-Type"), ";")[0]
		log.Println("req content type", contType)

		switch contType {
		case "multipart/form-data":
			a.handleFormData(ctx, userRole)
		case "application/json":
			a.handleApplicationJson(ctx, userRole)
		}
	}
}

// Parsing a slice of maps(rows of records) from csv data to a slice of participant structs
// NOTE: this function now handles db calls to append particiaptns to the database as well
func (a *App) ParseParticipants(db *gorm.DB, teamRecords []map[string]string) (*[]ParticipantInfo, error) {
	participants := []ParticipantInfo{}

	for i := 0; i < len(teamRecords); i++ {
		record := teamRecords[i]
		teamLeaderEmail := record["Email Address"]
		log.Println("team leader: ", teamLeaderEmail)

		team := database.Team{
			Team:  record["Team Name"],
			Email: teamLeaderEmail,
		}
		err := a.Store.CreateTeam(&team)
		if err != nil {
			log.Fatalln("failed to create a team", err)
			continue
		}

		for j := 1; j <= 5; j++ {
			participantName := record[fmt.Sprintf("Name %d", j)]
			if participantName == "" {
				continue
			} else {
				ph, err := strconv.Atoi(record[fmt.Sprintf("Phone %d", j)])
				if err != nil {
					return nil, err
				}

				participant := database.Participant{
					TeamID:    team.ID,
					Name:      strings.TrimSpace(record[fmt.Sprintf("Name %d", j)]),
					Phone:     int64(ph),
					Branch:    strings.TrimSpace(record[fmt.Sprintf("Branch %d", j)]),
					PesHostel: strings.TrimSpace(record[fmt.Sprintf("PES Hostel %d", j)]),
				}

				pid, _ := GenerateUUID(participant)
				signedString, _ := a.GenerateAuthoToken(participant, pid)

				_, err = a.GenerateQR(signedString, participant.Name, team.Team, teamLeaderEmail, pid)
				if err != nil {
					log.Fatal(err)
				}

				a.Store.CreateParticipant(&participant, database.CheckpointsWithDefaults(), pid, signedString)

				participantInfo := ParticipantInfo{
					participant,
					team,
				}

				participants = append(participants, participantInfo)
			}
		}
	}

	return &participants, nil
}

func (a *App) ParseVolunteers(db *gorm.DB, volunteerRecords []map[string]string) (*[]database.DBAuthorisedUsers, error) {
	volunteers := []database.DBAuthorisedUsers{}

	for _, record := range volunteerRecords {

		name := strings.TrimSpace(record["Name"])
		email := strings.TrimSpace(record["Email"])
		phoneStr := strings.TrimSpace(record["Phone"])

		phone, err := strconv.ParseInt(phoneStr, 10, 64)
		if err != nil {
			log.Println("Invalid phone number : ", phoneStr)
			return nil, err
		}

		volunteer := database.DBAuthorisedUsers{
			Name:          name,
			VerifiedEmail: email,
			Phone:         phone,
			UserRole:      "volunteer",
			SUB:           "",
		}

		err = a.Store.CreateVolunteer(&volunteer)
		if err != nil {
			log.Println("Error while storing", err)
		}

		volunteers = append(volunteers, volunteer)
	}

	return &volunteers, nil
}

func (a *App) ParseVolunteerManual(db *gorm.DB, name, email, phoneStr string) (int, *[]database.DBAuthorisedUsers, error) {
	volunteers := []database.DBAuthorisedUsers{}

	phone, err := strconv.ParseInt(phoneStr, 10, 64)
	if err != nil {
		log.Println("Invalid phone number : ", phoneStr)
		return 0, nil, err
	}

	volunteer := database.DBAuthorisedUsers{
		Name:          name,
		VerifiedEmail: email,
		Phone:         phone,
		UserRole:      "volunteer",
		SUB:           "",
	}

	success, err := a.Store.CreateVolunteerManual(&volunteer)
	if err != nil {
		log.Println("Error while storing", err)
	}

	volunteers = append(volunteers, volunteer)

	return success, &volunteers, nil
}

func (a *App) saveFile(file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}

func annotateQR(path string, name string) error {
	// no more than 40 characters allowed
	if len(name) > 40 {
		name = name[:40]
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	// skip non png files
	if format != "png" {
		log.Println("[Warning]: non png file in qr directory")
		return nil
	}

	dc := gg.NewContextForImage(img)

	dc.SetRGBA(0, 0, 0, 1)
	dc.DrawString(name, 10, 10)

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err = png.Encode(outFile, dc.Image()); err != nil {
		return err
	}

	return nil
}
