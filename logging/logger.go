package logging

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var log *zap.Logger

func InitLogger(moduleName, logFilePath string) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			zapcore.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(getLogFileWriter(logFilePath)),
			zapcore.InfoLevel,
		),
	)

	log = zap.New(core).Named(moduleName)
}

func getLogFileWriter(logFilePath string) zapcore.WriteSyncer {
	logDir := "./logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, os.ModePerm)
		if err != nil {
			return nil
		}
	}

	logFilePath = filepath.Join(logDir, fmt.Sprintf("%s_%s.log", logFilePath, time.Now().Format("20060102")))
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}
	return zapcore.AddSync(file)
}

func getModuleName() string {
	_, filename, _, _ := runtime.Caller(1)
	moduleName := filepath.Base(filepath.Dir(filename))
	return moduleName
}

func Info(msg string, tags ...zap.Field) {
	log.Info(msg, tags...)
}

func Warn(msg string, tags ...zap.Field) {
	log.Warn(msg, tags...)
}

func Error(msg string, tags ...zap.Field) {
	log.Error(msg, tags...)
}

func Debug(msg string, tags ...zap.Field) {
	log.Debug(msg, tags...)
}

func Sync() error {
	return log.Sync()
}

func ExecuteAndLogError(fn func() error) {
	if err := fn(); err != nil {
		Error("An error occurred while executing the function.", zap.Error(err))
	}
}
