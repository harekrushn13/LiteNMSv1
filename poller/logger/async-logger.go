package logger

import (
	"context"
	"go.uber.org/zap/zapcore"
)

type logItem struct {
	level zapcore.Level

	message string

	fields []zapcore.Field

	ctx context.Context
}

var (
	logChan = make(chan *logItem, 100)

	shutdownChan = make(chan struct{})
)

func startAsyncLogger() {

	go func() {

		for {

			select {

			case item := <-logChan:

				logByLevel(item.level, item.message, item.fields)

			case <-shutdownChan:

				return
			}
		}
	}()
}

func logByLevel(level zapcore.Level, msg string, fields []zapcore.Field) {

	switch level {

	case zapcore.DebugLevel:

		Logger.Debug(msg, fields...)

	case zapcore.InfoLevel:

		Logger.Info(msg, fields...)

	case zapcore.WarnLevel:

		Logger.Warn(msg, fields...)

	case zapcore.ErrorLevel:

		Logger.Error(msg, fields...)

	case zapcore.DPanicLevel:

		Logger.DPanic(msg, fields...)

	case zapcore.PanicLevel:

		Logger.Panic(msg, fields...)

	case zapcore.FatalLevel:

		Logger.Fatal(msg, fields...)

	}

}

func AsyncInfo(msg string, fields ...zapcore.Field) {

	logChan <- &logItem{level: zapcore.InfoLevel, message: msg, fields: fields}
}

func AsyncDebug(msg string, fields ...zapcore.Field) {

	logChan <- &logItem{level: zapcore.DebugLevel, message: msg, fields: fields}
}

func AsyncWarn(msg string, fields ...zapcore.Field) {

	logChan <- &logItem{level: zapcore.WarnLevel, message: msg, fields: fields}
}

func AsyncError(msg string, fields ...zapcore.Field) {

	logChan <- &logItem{level: zapcore.ErrorLevel, message: msg, fields: fields}
}

func StopAsyncLogger() {

	close(shutdownChan)
}
