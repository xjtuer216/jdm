package log

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	Logger *logrus.Logger
)

func Init(logDir string) error {
	Logger = logrus.New()

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, "jdm.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	Logger.SetOutput(f)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return nil
}

func GetLogger() *logrus.Logger {
	if Logger == nil {
		Logger = logrus.New()
	}
	return Logger
}