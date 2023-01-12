package logger

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func Init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func SetLevel(level string) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		log.Fatalf("Unable to log... wait what: %v\n", err)
	}
	log.SetLevel(logLevel)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		log.WithField("request_ip", clientIP).Debugf("Requesting %s", c.Request.URL)
		c.Next()
		requestLog := log.WithFields(log.Fields{
			"request_ip": clientIP, "status_code": c.Writer.Status(),
		})
		if c.Writer.Status() != http.StatusOK {
			requestLog.Errorf("Finished request %s", c.Request.URL)
			return
		}
		requestLog.Debugf("Finished request %s", c.Request.URL)
	}
}
