package main

import (
	"hot-coffee/internal/server"
	"hot-coffee/internal/setup"

	_ "github.com/lib/pq"
)

func main() {
	logger, logFile := setup.SetupLogger()
	defer logFile.Close()

	db := setup.SetupDatabase(logger)
	defer db.Close()

	server.ServerLaunch(db, logger)
}
