package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var Logger *zap.Logger

func InitLogger() error {

	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()

	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	fileEncoderConfig := zap.NewProductionEncoderConfig()

	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	consoleLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {

		return lvl == zapcore.InfoLevel
	})

	fileLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {

		return lvl != zapcore.InfoLevel
	})

	fileWriter := zapcore.AddSync(NewRotatingLogger(1, 7, 3, false))

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), consoleLevel)

	fileCore := zapcore.NewCore(fileEncoder, fileWriter, fileLevel)

	core := zapcore.NewTee(consoleCore, fileCore)

	Logger = zap.New(core, zap.AddCaller())

	return nil
}

func NewRotatingLogger(maxSizeMB, maxAgeDays, maxBackups int, compress bool) *lumberjack.Logger {

	return &lumberjack.Logger{

		Filename: "./logs/poller_log_" + time.Now().Format("2006-01-02") + ".log",

		MaxSize: maxSizeMB, // megabytes

		MaxAge: maxAgeDays, // days

		MaxBackups: maxBackups, // number of backups

		Compress: compress, // gzip old logs
	}
}
