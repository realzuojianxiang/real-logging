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

const (
	defaultLogDir      = "./logs"
	defaultLogFileName = "20060102"
)

var log *zap.Logger

func InitLogger(moduleName, logFilePath string) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	fileWriter, err := getLogFileWriter(logFilePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize log file writer: %v", err))
	}

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			zapcore.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			zapcore.InfoLevel,
		),
	)

	log = zap.New(core).Named(moduleName)
}

func getLogFileWriter(logFilePath string) (zapcore.WriteSyncer, error) {
	if _, err := os.Stat(defaultLogDir); os.IsNotExist(err) {
		err := os.Mkdir(defaultLogDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create log directory: %v", err)
		}
	}

	logFilePath = filepath.Join(defaultLogDir, fmt.Sprintf("%s_%s.log", logFilePath, time.Now().Format(defaultLogFileName)))
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	return zapcore.AddSync(file), nil
}

func appendCaller(fields []zap.Field) []zap.Field {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		fields = append(fields, zap.String("file", file), zap.Int("line", line))
	}
	return fields
}

func Info(msg string, tags ...zap.Field) {
	entry := log.Check(zapcore.InfoLevel, msg)
	if entry != nil {
		entry.Write(append(appendCaller(tags), zap.String("caller", getCaller()))...)
	}
}

func Warn(msg string, tags ...zap.Field) {
	entry := log.Check(zapcore.WarnLevel, msg)
	if entry != nil {
		entry.Write(append(appendCaller(tags), zap.String("caller", getCaller()))...)
	}
}

func Error(msg string, tags ...zap.Field) {
	entry := log.Check(zapcore.ErrorLevel, msg)
	if entry != nil {
		entry.Write(append(appendCaller(tags), zap.String("caller", getCaller()))...)
	}
}

func Debug(msg string, tags ...zap.Field) {
	entry := log.Check(zapcore.DebugLevel, msg)
	if entry != nil {
		entry.Write(append(appendCaller(tags), zap.String("caller", getCaller()))...)
	}
}

func getCaller() string {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return "unknown"
}

func Sync() error {
	return log.Sync()
}

func ExecuteAndLogError(fn func() error) {
	if err := fn(); err != nil {
		Error("An error occurred while executing the function.", zap.Error(err))
	}
}
