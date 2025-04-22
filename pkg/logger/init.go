package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func CreateLogger(name string) (*slog.Logger, error) {
	options := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	// r := &lumberjack.Logger{
	// 	Filename:   name,
	// 	LocalTime:  true,
	// 	MaxSize:    100, // megabytes
	// 	MaxAge:     28,  // days
	// 	MaxBackups: 3,
	// 	Compress:   false, // disabled by default
	// }
	// fileHandler := slog.NewTextHandler(r, options)
	// // 创建 Logger
	// logger := slog.New(fileHandler)
	// slog.SetDefault(logger)

	// 创建一个屏幕输出的 Handler
	consoleHandler := slog.NewTextHandler(os.Stdout, options)
	// 创建 Logger
	logger := slog.New(consoleHandler)
	slog.SetDefault(logger)

	return logger, nil
}

func init() {
	CreateLogger("log_rotate.txt")
}

type Level int

const (
	DebugLevel = slog.LevelDebug
	InfoLevel  = slog.LevelInfo
	WarnLevel  = slog.LevelWarn
	ErrorLevel = slog.LevelError
)

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

func Log(level Level, msg string, args ...any) {
	ctx := context.Background()
	logLevel := slog.Level(level)
	logObj := slog.Default()
	if !logObj.Enabled(ctx, logLevel) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(2, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), logLevel, msg, pc)
	r.Add(args...)
	_ = logObj.Handler().Handle(ctx, r)
}
