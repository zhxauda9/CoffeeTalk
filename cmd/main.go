package main

import (
	_ "github.com/lib/pq"
	"hot-coffee/internal/server"
	"hot-coffee/internal/setup"
)

func main() {
	logger, logFile := setup.SetupLogger()
	defer logFile.Close()

	db := setup.SetupDatabase(logger)
	defer db.Close()

	server.ServerLaunch(db, logger)
}
