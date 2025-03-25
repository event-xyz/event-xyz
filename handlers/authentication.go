package handlers

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/homebrew-ec-foss/eventloop/database"
	"github.com/skip2/go-qrcode"
)

type Env interface {
	Getenv(key string) string
}

type OsEnv struct{}

func (m *OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

type MockEnv struct {
	env map[string]string
}

type MockEnvPair struct {
	key   string
	value string
}

func (m *MockEnv) Getenv(key string) string {
	val, exists := m.env[key]
	if !exists {
		return ""
	}
	return val
}

func (m *MockEnv) Setenv(mockEnvPair MockEnvPair) {
	m.env[mockEnvPair.key] = mockEnvPair.value
}

// Claims for JWT
type JWTClaims struct {
	UUID  string `json:"UUID"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	// probably add team id here
	jwt.RegisteredClaims
}

var ErrJWTFailedClaimsParsing = fmt.Errorf("failed to parse for cliams. Seems like an invalid QR")

func (a *App) goDotEnvVariable(key string) string {
	return os.Getenv(key)
}

func (a *App) JWTAuthCheck(rawtoken string) (bool, *jwt.MapClaims) {
	parser_struct := jwt.Parser{}

	claims := jwt.MapClaims{}
	dotenv := a.goDotEnvVariable("JWT_SECRET_KEY")
	token, err := parser_struct.ParseWithClaims(rawtoken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(dotenv), nil
	})
	if err != nil {
		return false, nil
	}

	if token.Valid {
		// sucessful auth
		return true, &claims
	} else {
		return false, nil
	}
}

func GenerateUUID(user_record database.Participant) (string, error) {
	unique_string := fmt.Sprintf("%s", user_record.Name)

	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(unique_string))

	// handler error where generated UUID is nil
	return id.String(), nil
}

// BUG
//
//	All the jwt toekns use email in place of
//	college name :P
func (a *App) GenerateAuthoToken(user_record database.Participant, p_uuid string) (string, *JWTClaims) {
	// var id int64

	// TODO: generate UUI.
	claims := JWTClaims{
		// add a extra field for UUID
		p_uuid,
		user_record.Name,
		strconv.Itoa(int(user_record.Phone)),
		jwt.RegisteredClaims{
			Issuer: "userservices",
		},
	}

	dotenv := a.goDotEnvVariable("JWT_SECRET_KEY")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(dotenv))
	if err != nil {
		log.Println("Aiyo where is that []bytes")
		log.Fatal(err)
	}

	return signedString, &claims
}

func (a *App) GetClaimsInfo(rawtoken string) (map[string]interface{}, error) {
	parser_struct := jwt.Parser{}
	claims := jwt.MapClaims{}
	token, err := parser_struct.ParseWithClaims(rawtoken, claims, func(t *jwt.Token) (interface{}, error) {
		dotenv := a.goDotEnvVariable("JWT_SECRET_KEY")
		return []byte(dotenv), nil
	})
	if err != nil {
		log.Println(err)
		return nil, ErrJWTFailedClaimsParsing
	}

	if token.Valid {
		return claims, nil
	} else {
		log.Println(err)
		return nil, ErrJWTFailedClaimsParsing
	}
}

// TODO: cleanup arguments for GenerateQR
func GenerateQR(signedString, participantName string, leaderEmail string, uuid string) ([]byte, error) {
	var png []byte
	png, err := qrcode.Encode(signedString, qrcode.Low, 256)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll("../test-data/qr-png/", 0750); err != nil {
		log.Println(err)
	}

	err = qrcode.WriteFile(signedString, qrcode.Medium, 256, fmt.Sprintf("../test-data/qr-png/%s-%s-%s.png", leaderEmail, participantName, uuid))
	if err != nil {
		return nil, err
	}

	return png, nil
}

func (a *App) HandleParticipantSearch(ctx *gin.Context) {
    name := ctx.DefaultQuery("name", "")
    phone := ctx.DefaultQuery("phone", "")
    log.Printf("Searching participants with name: %s, phone: %s", name, phone)

    authHeader := ctx.GetHeader("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        log.Println("Missing or invalid Authorization header")
        ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Missing or invalid Authorization header"})
        return
    }
    jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

    valid, claims := a.JWTAuthCheck(jwtToken)
    if !valid {
        log.Println("Invalid JWT token")
        ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid JWT"})
        return
    }
    log.Printf("JWT claims: %v", claims)

    results, err := a.Store.SearchParticipants(name, phone)
    if err != nil {
        switch err {
        case database.ErrDbOpenFailure:
            log.Printf("Database error: %v", err)
            ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Database operation failed"})
            return
        case database.ErrDbMissingRecord:
            log.Println("No participants found")
            ctx.JSON(http.StatusNotFound, gin.H{"message": "No participants found"})
            return
        default:
            log.Printf("Unexpected error: %v", err)
            ctx.JSON(http.StatusInternalServerError, gin.H{"message": "An error occurred"})
            return
        }
    }

    log.Printf("Found %d participants", len(results))
    ctx.JSON(http.StatusOK, gin.H{
        "message":      "Participants fetched successfully",
        "participants": results,
    })
}