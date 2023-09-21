package services

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
	fields Fields
}

func NewLogrusLogger() Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	defaultFields := Fields{
		"service":  "crm-connector",
		"hostname": hostname,
	}

	return &LogrusLogger{
		logger: l,
		fields: defaultFields,
	}
}

func (l *LogrusLogger) WithFields(data Fields) Logger {
	cwd, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("Failed to determine working directory: %s", err)
	}
	formatter := time.Now().Format("2006-01-02")
	logPathLocation := filepath.Join(cwd, "/logfile/")
	if _, err := os.Stat(logPathLocation); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logPathLocation, os.ModePerm)
		if err != nil {
			logrus.Println(err)
		}
	}
	logFilePath := filepath.Join(cwd, "/logfile/", formatter+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logrus.Fatalf("Failed to open log file %s for output: %s", logFilePath, err)
	}

	l.logger.SetOutput(io.MultiWriter(os.Stderr, logFile))
	return &LogrusLogger{
		logger: l.logger,
		fields: data,
	}
}

func (l *LogrusLogger) Debug(msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).Debug(msg)
}

func (l *LogrusLogger) Debugf(msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Debugf(msg, args...)
}

func (l *LogrusLogger) Info(msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).Info(msg)
}

func (l *LogrusLogger) Infof(msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Infof(msg, args...)
}

func (l *LogrusLogger) Warn(msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).Warn(msg)
}

func (l *LogrusLogger) Warnf(msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Warnf(msg, args...)
}

func (l *LogrusLogger) Error(err error, msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Error(msg)
}

func (l *LogrusLogger) Errorf(err error, msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Errorf(msg, args...)
}

func (l *LogrusLogger) Fatal(err error, msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Fatal(msg)
}

func (l *LogrusLogger) Fatalf(err error, msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Fatalf(msg, args...)
}

func (l *LogrusLogger) Printf(format string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Printf(format, args)
}

func (l *LogrusLogger) Println(args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Println(args)
}
