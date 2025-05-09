package logger

import (
	"context"
	"fmt"
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
	showLogLevel               = InfoLevel
)

func init() {
	var err error
	showLogLevel, err = GetLevel(os.Getenv("PIE_LOG_LEVEL"))
	if err != nil {
		showLogLevel = InfoLevel
	}

	options := &slog.HandlerOptions{
		AddSource: showSource,
		Level:     slog.Level(showLogLevel),
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

	if err != nil {
		Log(ErrorLevel, "init log", "error", err.Error())
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	}
	return ""
}

func GetLevel(levelStr string) (Level, error) {
	switch levelStr {
	case DebugLevel.String():
		return DebugLevel, nil
	case InfoLevel.String():
		return InfoLevel, nil
	case WarnLevel.String():
		return WarnLevel, nil
	case ErrorLevel.String():
		return ErrorLevel, nil
	}
	return InfoLevel, fmt.Errorf("unknown Level String: '%s', defaulting to InfoLevel", levelStr)
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
