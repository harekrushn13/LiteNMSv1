package logger

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

func NewLogger(logDir string, maxSizeMB, maxAge, maxBackups int, compress bool) *lumberjack.Logger {

	filename := filepath.Join(logDir, "poller_log.log")

	info, err := os.Stat(filename)

	if err == nil && info.Size() > int64(maxSizeMB*1024*1024) {

		rotatedName := info.Name() + "-" + time.Now().Format("2006-01-02T15-04-05") + ".log"

		_ = os.Rename(filename, rotatedName)
	}

	return &lumberjack.Logger{
		Filename: filename,

		MaxSize: maxSizeMB,

		MaxAge: maxAge,

		MaxBackups: maxBackups,

		Compress: compress,
	}
}
