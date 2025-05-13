package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
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

	fileWriter := zapcore.AddSync(NewLogger("logs", 1, 7, 3, false))

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), consoleLevel)

	fileCore := zapcore.NewCore(fileEncoder, fileWriter, fileLevel)

	core := zapcore.NewTee(consoleCore, fileCore)

	Logger = zap.New(core, zap.AddCaller())

	startAsyncLogger()

	return nil
}
