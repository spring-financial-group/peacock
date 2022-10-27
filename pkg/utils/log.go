package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// InitLogger initialises the logger with Peacock settings
func InitLogger(isVerbose bool) {
	logLevel := log.InfoLevel
	if isVerbose {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		QuoteEmptyFields: true,
	})
	log.SetOutput(os.Stdout)
}
