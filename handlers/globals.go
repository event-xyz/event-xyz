package handlers

import (
	"log"
	"os"

	"github.com/homebrew-ec-foss/eventloop/database"
)

type App struct {
	Store  *database.EventDB
	Logger *log.Logger
}

func InitializeAppWithConfig(dbPath string) App {
	db, err := database.TryInitializeDB(dbPath)

	if err != nil {
		log.Fatalln("[MAIN] failed to run InitializeDB")
	}

	return App{
		Store:  db,
		Logger: log.New(os.Stderr, "LOG\t", log.Ldate|log.Ltime|log.Lshortfile),
	}
}
