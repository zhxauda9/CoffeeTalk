package setup

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func SetupDatabase(logger *slog.Logger) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Error("Error creating sql.DB", "Error", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		logger.Error("Error connecting to the database", "Error", err)

		for i := 1; i <= 5; i++ {
			logger.Info(fmt.Sprintf("Reconnecting... Attempt #%d", i))
			time.Sleep(3 * time.Second)

			if err = db.Ping(); err == nil {
				logger.Info("Successful connection to the database!")
				return db
			}
		}

		logger.Error("Failed to connect to the database after 5 attempts. Completing the work.")
		os.Exit(1)
	}

	return db
}
