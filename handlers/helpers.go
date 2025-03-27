package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/homebrew-ec-foss/eventloop/database"
	"gorm.io/gorm"
)

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
	}

	user_sub := string(content.Bytes())
	log.Println("sub:", user_sub)

	_, err = a.Store.SubAuthentication(user_sub, "admin")
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

		contType := ctx.Request.Header.Get("Content-Type")
		contType = strings.Split(contType, ";")[0]
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
func (a *App) ParseParticipants(db *gorm.DB, teamRecords []map[string]string) (*[]database.Participant, error) {
	participants := []database.Participant{}

	for i := 0; i < len(teamRecords); i++ {
		record := teamRecords[i]
		teamLeaderEmail := record["Email"]

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

				_, err = GenerateQR(signedString, participant.Name, teamLeaderEmail, pid)
				if err != nil {
					log.Fatal(err)
				}

				a.Store.CreateParticipant(&participant, database.CheckpointsWithDefaults(), pid, signedString)

				participants = append(participants, participant)
			}
		}
	}

	return &participants, nil
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
