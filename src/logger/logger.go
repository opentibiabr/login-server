package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var logger = log.New()
var logFile *os.File
var loggerMu sync.Mutex

type Entry struct {
	fields log.Fields
}

func Init(level log.Level, logFilePaths ...string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	closeLogFile()
	logger.SetLevel(level)
	logger.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
	logger.SetOutput(os.Stdout)

	if len(logFilePaths) == 0 || logFilePaths[0] == "" {
		return
	}

	file, err := openLogFile(logFilePaths[0])
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to open log file %q: %s", logFilePaths[0], err.Error()))
		return
	}

	logFile = file
	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func WithFields(fields log.Fields) *Entry {
	copiedFields := log.Fields{}
	for key, value := range fields {
		copiedFields[key] = value
	}

	return &Entry{fields: copiedFields}
}

func (entry *Entry) Debug(message string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.WithFields(entry.fields).Debug(message)
}

func (entry *Entry) Info(message string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.WithFields(entry.fields).Info(message)
}

func Debug(message string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.Debug(message)
}

func Info(message string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.Info(message)
}

func Warn(message string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.Warn(message)
}

func Error(err error) {
	if err == nil {
		return
	}

	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.Error(err.Error())
}

func Panic(err error) {
	if err == nil {
		return
	}

	loggerMu.Lock()
	defer loggerMu.Unlock()

	logger.Panic(err.Error())
}

func LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		WithFields(log.Fields{
			"0": c.Writer.Status(),
			"1": "web-server",
			"2": fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
			"3": c.ClientIP(),
			"4": c.Request.Method,
		}).Info(c.Request.URL.Path)
	}
}

func openLogFile(path string) (*os.File, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

func closeLogFile() {
	if logFile == nil {
		return
	}

	_ = logFile.Close()
	logFile = nil
}
