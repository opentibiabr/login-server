package logger

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var logger = log.New()

func Init(level log.Level) {
	logger.SetLevel(level)
	logger.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
}

func LogRequest(code int, payload interface{}, message string, fields log.Fields) {
	fields["1"] = code
	logger.WithFields(fields).Info(message)

	fields["5"] = payload
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

func Fatal(err error) {
	logger.Error(err.Error())
}

func BuildRequestLogFields(r *http.Request, start time.Time) log.Fields {
	return log.Fields{
		"2": "web-server",
		"3": fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
		"4": fmt.Sprintf("%s %s", r.Method, r.URL),
	}
}
