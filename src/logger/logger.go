package logger

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var logger = log.New()

func Init(level log.Level) {
	logger.SetLevel(level)
	logger.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
}

func LogRequest(r *http.Request, code int, payload interface{}, message string) {
	fields := log.Fields{
		"component": "web-server",
		"code":      code,
		"url":       fmt.Sprintf("%s %s", r.Method, r.URL),
	}

	logger.WithFields(fields).Info(message)

	fields["payload"] = payload
	logger.WithFields(fields).Debug()
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
