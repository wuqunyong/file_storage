package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Level int

const (
	DebugLevel = Level(slog.LevelDebug)
	InfoLevel  = Level(slog.LevelInfo)
	WarnLevel  = Level(slog.LevelWarn)
	ErrorLevel = Level(slog.LevelError)
)

var (
	consoleLogger *slog.Logger = nil
	fileLogger    *slog.Logger = nil
	showConsole                = true
	showSource                 = true
	showLogLevel               = slog.LevelInfo
)

func init() {
	options := &slog.HandlerOptions{
		AddSource: showSource,
		Level:     showLogLevel,
	}
	// 创建一个屏幕输出的 Handler
	consoleHandler := slog.NewTextHandler(os.Stdout, options)
	consoleLogger = slog.New(consoleHandler)

	// 创建一个文件输出的 Handler
	name := "log_rotate.txt"
	r := &lumberjack.Logger{
		Filename:   name,
		LocalTime:  true,
		MaxSize:    100, // megabytes
		MaxAge:     28,  // days
		MaxBackups: 3,
		Compress:   false, // disabled by default
	}
	fileHandler := slog.NewTextHandler(r, options)
	fileLogger = slog.New(fileHandler)
}

func Log(level Level, msg string, args ...any) {
	ctx := context.Background()
	logLevel := slog.Level(level)

	if fileLogger == nil {
		panic("fileLogger not init")
	}

	if !fileLogger.Enabled(ctx, logLevel) {
		return
	}

	var pc uintptr
	if showSource {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(2, pcs[:])
		pc = pcs[0]
	}

	r := slog.NewRecord(time.Now(), logLevel, msg, pc)
	r.Add(args...)
	_ = fileLogger.Handler().Handle(ctx, r)

	if showConsole {
		if consoleLogger == nil {
			panic("consoleLogger not init")
		}
		consoleLogger.Handler().Handle(ctx, r)
	}
}
