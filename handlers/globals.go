package handlers

import (
	"log"
	"os"
	"time"

	"github.com/homebrew-ec-foss/eventloop/database"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	EnvPath string `yaml:"envPath"`
	DbPath  string `yaml:"dbPath"`
	QrPath  string `yaml:"qrPath"`
}

type App struct {
	Store    *database.EventDB
	Logger   *log.Logger
	Config   Config
	Env      Env
	Metrics  *PrometheusMetrics
	AppStart time.Time
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

	app := &App{
		Store:    db,
		Logger:   log.New(os.Stderr, "LOG\t", log.Ldate|log.Ltime|log.Lshortfile),
		Config:   Config{
			DbPath: dbpath,
		},
		Env:      &mockEnv,
		Metrics:  prom(),
		AppStart: time.Now(),
	}

	app.InitializeMetrics()
	return app
}

func InitializeAppWithConfig(configPath string) *App {
	var buf []byte
	var config Config

	buf, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln("failed to load config file: ", err)
	}

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		log.Fatalln(err)
	}

	if config.QrPath == "" {
		config.QrPath = "../shared/"
	}

	err = godotenv.Load(config.EnvPath)
	if err != nil {
		log.Fatalln("failed to open env file: ", err)
	}

	db, err := database.TryInitializeDB(config.DbPath)
	if err != nil {
		log.Fatalln("[MAIN] failed to run InitializeDB")
	}

	app := &App{
		Store:    db,
		Logger:   log.New(os.Stderr, "LOG\t", log.Ldate|log.Ltime|log.Lshortfile),
		Config:   config,
		Env:      &OsEnv{},
		Metrics:  prom(),
		AppStart: time.Now(),
	}

	app.InitializeMetrics()
	return app
}
