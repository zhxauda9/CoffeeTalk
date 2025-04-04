package setup

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

func SetupLogger() (*slog.Logger, *os.File) {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0o755)
	}

	logFileName := fmt.Sprintf("logs/log_%s.log", time.Now().Format("20060102_150405"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Printf("Error creating the logs file: %v\n", err)
		os.Exit(1)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(multiWriter, nil))

	logger.Info("Server starting...", "logFile", logFileName)
	return logger, logFile
}
