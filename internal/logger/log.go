package logger

import (
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

const (
	out_type_stderr = "stderr"
	out_type_file   = "file"
)

type Log struct {
	Logger   *logrus.Logger
	outType  string
	pathfile string
}

func NewLogger() *Log {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(os.Stderr)

	return &Log{Logger: logger, outType: out_type_stderr}
}

func NewLoggerWithFile(pathfile string) *Log {
	if _, err := os.Stat(pathfile); os.IsNotExist(err) {
		err := os.MkdirAll(path.Dir(pathfile), os.ModePerm)
		if err != nil {
			logrus.Error(err)
		}
	}

	f, err := os.OpenFile(pathfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logrus.Error(fmt.Sprintf("error opening file: %v", err.Error()))
	}
	defer f.Close()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(f)

	return &Log{Logger: logger, outType: out_type_file, pathfile: pathfile}
}

func openFile(pathfile string) *os.File {
	f, err := os.OpenFile(pathfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logrus.Error("error opening file: %v", err.Error())
		return nil
	}
	return f
}

func (logger *Log) Info(msg interface{}, fields map[string]interface{}) {
	switch logger.outType {
	case out_type_file:
		if f := openFile(logger.pathfile); f != nil {
			defer f.Close()
			logger.Logger.SetOutput(f)
			if fields != nil {
				logger.Logger.WithFields(fields).Info(msg)
			} else {
				logger.Logger.Info(msg)
			}
		}
	case out_type_stderr:
		if fields != nil {
			logger.Logger.WithFields(fields).Info(msg)
		} else {
			logger.Logger.Info(msg)
		}
	}
}

func (logger *Log) Error(msg interface{}, fields map[string]interface{}) {
	switch logger.outType {
	case out_type_file:
		if f := openFile(logger.pathfile); f != nil {
			defer f.Close()
			logger.Logger.SetOutput(f)
			if fields != nil {
				logger.Logger.WithFields(fields).Error(msg)
			} else {
				logger.Logger.Error(msg)
			}
		}
	case out_type_stderr:
		if fields != nil {
			logger.Logger.WithFields(fields).Error(msg)
		} else {
			logger.Logger.Error(msg)
		}
	}
}
