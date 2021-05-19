package logger

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

var logger = log.New()

func Init(level log.Level) {
	logger.SetLevel(level)
	logger.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
}

func WithFields(fields log.Fields) *log.Entry {
	return logger.WithFields(fields)
}

func Debug(message string) {
	logger.Debug(message)
}

func Info(message string) {
	logger.Info(message)
}

func Warn(message string) {
	logger.Warn(message)
}

func Error(err error) {
	logger.Error(err.Error())
}

func Panic(err error) {
	logger.Panic(err.Error())
}

func LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.WithFields(log.Fields{
			"0": c.Writer.Status(),
			"1": "web-server",
			"2": fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
			"3": c.ClientIP(),
			"4": c.Request.Method,
		}).Info(c.Request.URL.Path)
	}
}
