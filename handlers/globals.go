package handlers

import (
	"encoding/json"
	"log"
	"os"

	"github.com/homebrew-ec-foss/eventloop/database"
	"github.com/joho/godotenv"
)

type Config struct {
	EnvPath string
	DbPath  string
	QrPath  string
}

type App struct {
	Store  *database.EventDB
	Logger *log.Logger
	Config Config
	Env    Env
}

func InitializeMockApp(dbpath string, secretPair MockEnvPair) *App {
	db, err := database.TryInitializeDB(dbpath)
	if err != nil {
		log.Fatalln(err)
	}
	mockEnv := MockEnv{
		env: make(map[string]string),
	}
	mockEnv.Setenv(secretPair)

	return &App{
		Store:  db,
		Logger: log.New(os.Stderr, "LOG\t", log.Ldate|log.Ltime|log.Lshortfile),
		Config: Config{
			DbPath: dbpath,
		},
		Env: &mockEnv,
	}
}

func InitializeAppWithConfig(configPath string) *App {
	var buf []byte
	var config Config

	buf, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Fatalln(err)
	}

	if config.QrPath == "" {
		config.QrPath = "../shared/"
	}

	err = godotenv.Load(config.EnvPath)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := database.TryInitializeDB(config.DbPath)
	if err != nil {
		log.Fatalln("[MAIN] failed to run InitializeDB")
	}

	return &App{
		Store:  db,
		Logger: log.New(os.Stderr, "LOG\t", log.Ldate|log.Ltime|log.Lshortfile),
		Config: config,
		Env:    &OsEnv{},
	}
}
