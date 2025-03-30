package handlers

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/homebrew-ec-foss/eventloop/database"
)

func (a *App) HandleCreateTest(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.String(
			http.StatusBadRequest,
			"Error: No file uploaded",
		)
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to open file",
		)
		return
	}

	defer fileContent.Close()

	reader := bufio.NewReader(fileContent)
	content := bytes.Buffer{}

	_, err = io.Copy(&content, reader)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to read file",
		)
		return
	}

	csvReader := csv.NewReader(bytes.NewReader(content.Bytes()))
	formData, err := csvReader.ReadAll()

	formHeaders := formData[0]
	formEntriesMap := make([]map[string]string, 0)

	for i := 1; i < len(formData); i++ {
		entry := make(map[string]string)
		for j := 0; j < len(formHeaders); j++ {
			entry[formHeaders[j]] = formData[i][j]
		}
		formEntriesMap = append(formEntriesMap, entry)
	}

	participants, err := a.ParseParticipants(a.Store.Db, formEntriesMap)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to write records to database",
		)
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{"data": participants},
	)
}

// TODO:
// HandleCreate only handles the incoming file
// - Create db based on event name
// - Store imporatnt event related info regarding date, etc...
func (a *App) HandleCreate(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.String(
			http.StatusBadRequest,
			"Error: No file uploaded",
		)
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to open file",
		)
		return
	}
	defer fileContent.Close()

	reader := bufio.NewReader(fileContent)
	content := bytes.Buffer{}
	_, err = io.Copy(&content, reader)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to read file",
		)
		return
	}

	// Reading csv to a 2-D slice
	csvReader := csv.NewReader(bytes.NewReader(content.Bytes()))
	formData, err := csvReader.ReadAll()

	// TODO(FUTURE): do proper checks here for form headers
	// check for validation of opinionated headers and
	// dynamic headers
	formHeaders := formData[0]
	formEntriesMap := make([]map[string]string, 0)

	// Converting csv data to a slice of maps (slice has several rows of records, where each row is a map)
	// Each map contains key-value pairs, where the key is the csv-header for the column
	for i := 1; i < len(formData); i++ {
		entry := make(map[string]string)
		for j := 0; j < len(formHeaders); j++ {
			entry[formHeaders[j]] = formData[i][j]
		}
		// Appending map(row) to slice(all rows)
		formEntriesMap = append(formEntriesMap, entry)
	}

	// Parsing the csv to a slice of Participants struct
	participants, err := a.ParseParticipants(a.Store.Db, formEntriesMap)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to write records to the database",
		)
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{"data": participants},
	)
}

// Handler receives the JWT and manages checkpoint(Eg: Dinner) updates
func (a *App) HandleCheckpoint(ctx *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"Error": "Failed to read request body"},
		)
		return
	}
	log.Println("QR code content received:", string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"Error": "Failed to parse JSON"},
		)
		return
	}

	jwtClaims, err := a.GetClaimsInfo(data["jwt"].(string))
	if err != nil && jwtClaims == nil {
		log.Println("Invalid jwt")
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Failed to pasrse JWT for the cliams. Seems like an invlaid QR"},
		)
		return
	}

	checkpointName := data["checkpoint"].(string)

	if jwtClaims == nil {
		log.Println("Invalid JWT")
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"Error": "Invalid JWT"},
		)
		return
	}

	dbParticipant, checkpointCleared, err := a.Store.ParticipantCheckpoint(jwtClaims["UUID"].(string), checkpointName)

	switch err {
	case database.ErrCheckpointCrossed:
		{
			log.Println(err)
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "The participant has already crossed the checkpoint"},
			)
			return
		}
	case database.ErrDbOpenFailure:
		{
			log.Println(err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"message": "Database OP failed server side, contact operators"},
			)
			return
		}
	case database.ErrDbMissingRecord:
		{
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"message":           "The QR might not be accurate",
					"checkpointCleared": false,
					"operation":         true,
				},
			)
			return
		}
	case database.ErrParticipantAbsent:
		{
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "Participant has never checked into the envet. Can't proceed with operation"},
			)
			return
		}
	case database.ErrParticipantLeft:
		{
			log.Println("Already left the event")
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "Participant has left the event. Can't proceed with operation"},
			)
			return
		}
	}

	// Respond to the client
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"message":           "QR code parsed sucessfully and operation sucessful",
			"checkpointCleared": checkpointCleared,
			"operation":         true,
			"dbParticipant":     dbParticipant,
		},
	)
}

// Handler receives the JWT and manages event enrty updates
func (a *App) HandleCheckin(ctx *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println(err)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Failed to read request body"},
		)
		return
	}
	log.Println("QR code content received:", string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to parse JSON"},
		)
		return
	}

	// Parsing claims and validating JWT
	jwtClaims, err := a.GetClaimsInfo(data["jwt"].(string))
	if err != nil && jwtClaims == nil {
		log.Println("Invalid jwt")
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Failed to pasrse JWT for the cliams. Seems like an invlaid QR"},
		)
		return
	}

	// Querying DB for participant and updating with entry
	dbParticipant, checkin, err := a.Store.ParticipantEntry(jwtClaims["UUID"].(string))

	switch err {
	case database.ErrDbOpenFailure:
		{
			log.Println(err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"message": "Database OP failed server side, contact operators"},
			)
			return
		}
	case database.ErrDbMissingRecord:
		{
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"message":   "The QR might not be accurate",
					"checkin":   false,
					"operation": true,
				},
			)
			return
		}
	case database.ErrParticipantLeft:
		{
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "Participant has left the event. Can't proceed with operation"},
			)
			return
		}
	}

	// Respond to the client
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"message":       "QR JWT parsed and db operation was sucessful",
			"checkin":       checkin,
			"operation":     true,
			"dbParticipant": dbParticipant,
		},
	)
	return
}

func (a *App) HandleCheckout(ctx *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
		return
	}

	// DEBUG

	// Parsing claims and validating JWT
	jwtClaims, err := a.GetClaimsInfo(data["jwt"].(string))

	if err != nil && jwtClaims == nil {
		log.Println("Invalid jwt")
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Failed to pasrse JWT for the cliams. Seems like an invlaid QR"},
		)
		return
	}

	log.Println(jwtClaims)

	// Querying DB for participant and updating with entry
	dbParticipant, checkout, err := a.Store.ParticipantExit(jwtClaims["UUID"].(string))

	switch err {
	case database.ErrDbOpenFailure:
		{
			log.Println(err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"message": "Database OP failed server side, contact operators"},
			)
			return
		}
	case database.ErrDbMissingRecord:
		{
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "The QR might not be accurate", "checkin": false, "operation": true},
			)
			return
		}
	case database.ErrParticipantAbsent:
		{
			log.Println("The guy never came!")
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "Participant has never checked into the envet. Can't proceed with operation"},
			)
			return
		}
	case database.ErrParticipantLeft:
		{
			log.Println("The guy left off -_-!")
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "Participant has already left the event. Can't proceed with operation"},
			)
			return
		}
	}

	// Respond to the client
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"message":       "QR JWT parsed and db operation was sucessful",
			"checkout":      checkout,
			"operation":     true,
			"dbParticipant": dbParticipant,
		},
	)
}

func (a *App) HandleParticipantSearch(ctx *gin.Context) {}

func (a *App) HandleQRFetch(ctx *gin.Context) {
	log.Println("Recieved")
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println(err)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Failed to read request body"},
		)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to parse JSON"},
		)
		return
	}

	// Parsing claims and validating JWT
	jwtClaims, err := a.GetClaimsInfo(data["jwt"].(string))

	if err != nil && jwtClaims == nil {
		log.Println("Invalid jwt")
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Failed to pasrse JWT for the cliams. Seems like an invlaid QR"},
		)
		return
	}

	dbparticipant, err := a.Store.JWTFetchParticipant(data["jwt"].(string))

	log.Println(dbparticipant)

	if errors.Is(err, database.ErrDbOpenFailure) {
		log.Println(err)
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"message":       "details parseed sucessfully",
			"dbParticipant": dbparticipant},
	)
}

// TODO:
// This endpoint exposes way too much data
// Non JWTID basesd search expects the name and phone to
// be unique together
func (a *App) HandleParticipantFetch(ctx *gin.Context) {
	jwtID := ctx.DefaultQuery("jwtID", "")
	partName := ctx.DefaultQuery("pname", "")
	partPhone := ctx.DefaultQuery("pphone", "")

	fmt.Println(jwtID, partName, partPhone)

	if jwtID != "" {
		// search based on JWT ID
		valid, _ := a.JWTAuthCheck(jwtID)
		if !valid {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"message": "Invalid jwt ID"},
			)
			return
		}
		dbParticipant, err := a.Store.JWTFetchParticipant(jwtID)
		if errors.Is(err, database.ErrDbOpenFailure) {
			log.Println(err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"message": "Database OP failed server side, contact operators"},
			)
			return
		}
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message":       "participant fetched successfully",
				"dbParticipant": dbParticipant},
		)
		return
	} else {
		log.Println("Fetching based on ID")
		dbParticipant, err := a.Store.FetchParticipant(partName, partPhone)
		switch err {
		case database.ErrDbOpenFailure:
			{
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"message": "Database OP failed server side, contact operators"},
				)
				return
			}
		case database.ErrDbMissingRecord:
			{
				ctx.JSON(
					http.StatusBadRequest,
					gin.H{"message": "No participant exists. Details may be incorrect"},
				)
				return
			}
		}
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message":       "participant fetched successfully",
				"dbParticipant": dbParticipant},
		)
		return
	}
}

func (a *App) HandleLogin(ctx *gin.Context) {
	// apparently map only i needto do
	var requestBody map[string]interface{}

	// json-> map tried doing with string because ez but was not nice
	if err := ctx.BindJSON(&requestBody); err != nil {
		log.Println("Error binding JSON:", err)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Invalid request"},
		)
		return
	}

	// print ingo
	log.Println("Received login request:", requestBody)

	// dbAuthUser, err := database.VerifyLogin(int64(strconv.Itoa(requestBody["sub"].(string))),requestBody["email"].(string))

	incomingUserReq := database.DBAuthorisedUsers{
		VerifiedEmail: requestBody["email"].(string),
		SUB:           requestBody["sub"].(string),
	}

	dbAuthUser, err := a.Store.VerifyLogin(incomingUserReq)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Database OP failed server side, contact operators",
				"success": false,
			},
		)
	} else {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message":    "Login successful",
				"success":    true,
				"dbAuthUser": dbAuthUser,
			},
		)
	}
	// if err == nil {
	// 	log.Println("there is NO error")
	// }
	//
	// switch err {
	// case database.ErrDbOpenFailure:
	// 	{
	// 		log.Println("couldnt return user")
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Database OP failed server side, contact operators", "success": false})
	// 		return
	// 	}
	// case database.ErrDbMissingRecord:
	// 	{
	// 		log.Println("Missing databse record")
	// 		ctx.JSON(http.StatusBadRequest, gin.H{"message": "No records available for", "success": false})
	// 		return
	// 	}
	// }

	// Respond to the client
	// ctx.JSON(http.StatusOK, gin.H{"message": "Login successful", "success": true, "dbAuthUser": dbAuthUser})
}

func HandleParticipantUpdate(ctx *gin.Context) {
	ctx.JSON(
		http.StatusOK,
		gin.H{"message": "Participants details updated sucessfully"},
	)
}

func (a *App) HandleVolunteers(ctx *gin.Context) {
	method := ctx.Param("method")

	switch method {
	case "csv":
		a.HandleVolunteersCsv(ctx)
	case "manual":
		a.HandleVolunteerManual(ctx)
	default:
		ctx.String(
			http.StatusBadRequest,
			"Error: No method specified",
		)
	}
}

func (a *App) HandleVolunteersCsv(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.String(
			http.StatusBadRequest,
			"Error: No file uploaded",
		)
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to open file",
		)
		return
	}
	defer fileContent.Close()

	reader := bufio.NewReader(fileContent)
	content := bytes.Buffer{}
	_, err = io.Copy(&content, reader)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to read file",
		)
		return
	}

	// Reading csv to a 2-D slice
	csvReader := csv.NewReader(bytes.NewReader(content.Bytes()))
	formData, err := csvReader.ReadAll()

	formHeaders := formData[0]
	formEntriesMap := make([]map[string]string, 0)

	// Converting csv data to a slice of maps (slice has several rows of records, where each row is a map)
	// Each map contains key-value pairs, where the key is the csv-header for the column
	for i := 1; i < len(formData); i++ {
		entry := make(map[string]string)
		for j := 0; j < len(formHeaders); j++ {
			entry[formHeaders[j]] = formData[i][j]
		}
		// Appending map(row) to slice(all rows)
		formEntriesMap = append(formEntriesMap, entry)
	}

	// Parsing the csv to a slice of Participants struct
	volunteers, err := a.ParseVolunteers(a.Store.Db, formEntriesMap)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to write records to the database",
		)
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{"data": volunteers},
	)
}

func (a *App) HandleVolunteerManual(ctx *gin.Context) {
	name := ctx.PostForm("name")
	email := ctx.PostForm("email")
	phone := ctx.PostForm("phone")

	log.Println(name, email, phone)
	// Parse volunteer manually.
	success, volunteer, err := a.ParseVolunteerManual(a.Store.Db, name, email, phone)
	if err != nil {
		ctx.String(
			http.StatusInternalServerError,
			"Error: Failed to write records to the database",
		)
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data":            volunteer,
			"deletionSuccess": success,
		},
	)
}
