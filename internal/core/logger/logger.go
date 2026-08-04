package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

var key = loggerContextKey{}

type Logger struct {
	*zap.Logger

	file *os.File
}

func NewLogger(config Config) (*Logger, error) {
	const formatTimestamp = "2006-01-02T15-04-05.000000"

	zapLvl := zap.NewAtomicLevel()

	if err := zapLvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	if err := os.MkdirAll(config.Folder, 0755); err != nil {
		return nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format(formatTimestamp)
	logFilePath := filepath.Join(
		config.Folder,
		fmt.Sprintf("%s.log", timestamp),
	)

	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	zapConfig := zap.NewDevelopmentEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout(formatTimestamp)

	zapEncoder := zapcore.NewConsoleEncoder(zapConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(zapEncoder, zapcore.Lock(logFile), zapLvl),
		zapcore.NewCore(zapEncoder, zapcore.Lock(os.Stdout), zapLvl),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return &Logger{
		Logger: logger,
		file:   logFile,
	}, nil
}

func (l *Logger) Close() {
	if err := l.Sync(); err != nil {
		fmt.Println("sync logger:", err)
	}

	if err := l.file.Close(); err != nil {
		fmt.Println("close log file: ", err)
	}
}

func (l *Logger) With(field ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(field...),
		file:   l.file,
	}
}

func ToContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(
		ctx,
		key,
		log,
	)
}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(key).(*Logger)
	if !ok {
		panic("no log in context")
	}

	return log
}
