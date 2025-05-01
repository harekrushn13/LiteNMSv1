package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger *zap.Logger

func InitLogger() error {

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	consoleCore := zapcore.NewCore(

		consoleEncoder,

		zapcore.AddSync(os.Stdout),

		zapcore.InfoLevel,
	)

	file, err := os.OpenFile("poller_errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {

		return err
	}

	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	fileCore := zapcore.NewCore(

		fileEncoder,

		zapcore.AddSync(file),

		zapcore.WarnLevel,
	)

	core := zapcore.NewTee(consoleCore, fileCore)

	Logger = zap.New(core, zap.AddCaller())

	return nil
}
