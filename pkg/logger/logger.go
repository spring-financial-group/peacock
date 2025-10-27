package logger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init() {
	// zerolog outputs JSON by default, no need to set formatter
	// Set global log level to debug
	Logger = log.Logger.Level(zerolog.DebugLevel)
}

func SetLevel(level string) {
	switch level {
	case "debug":
		Logger = log.Logger.Level(zerolog.DebugLevel)
	case "info":
		Logger = log.Logger.Level(zerolog.InfoLevel)
	case "warn":
		Logger = log.Logger.Level(zerolog.WarnLevel)
	case "error":
		Logger = log.Logger.Level(zerolog.ErrorLevel)
	default:
		log.Fatal().Msgf("Unable to log... wait what: %v", level)
	}
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		log.Debug().Str("request_ip", clientIP).Msgf("Requesting %s", c.Request.URL)
		c.Next()

		logger := log.With().Str("request_ip", clientIP).Int("status_code", c.Writer.Status()).Logger()
		if c.Writer.Status() != http.StatusOK {
			logger.Error().Msgf("Finished request %s", c.Request.URL)
			return
		}
		logger.Debug().Msgf("Finished request %s", c.Request.URL)
	}
}
