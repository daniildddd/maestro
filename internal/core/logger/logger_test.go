package logger_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
)

//nolint:paralleltest // NewLogger writes to process-wide os.Stdout (hardcoded)
func TestNewLogger(t *testing.T) {
	t.Run("valid config returns logger and creates log file", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)
		dir := t.TempDir()
		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: dir,
		})
		must.NoError(err)
		must.NotNil(log)

		defer func() {
			is.NoError(log.Close(), "logger close should not fail")
		}()

		log.Info("flushed message")

		entries, err := os.ReadDir(dir)
		must.NoError(err)

		must.Len(entries, 1)
		must.Contains(entries[0].Name(), ".log")
		data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
		must.NoError(err)

		is.Contains(string(data), "flushed message")
	})

	t.Run("invalid level returns error", func(t *testing.T) {
		must := require.New(t)
		dir := t.TempDir()
		log, err := logger.NewLogger(logger.Config{
			Level:  "INVALID_LEVEL",
			Folder: dir,
		})
		must.Error(err)
		must.Nil(log)
		must.ErrorContains(err, "unmarshal log level")
	})

	t.Run("non-creatable folder returns error", func(t *testing.T) {
		must := require.New(t)
		dir := t.TempDir()

		filePath := filepath.Join(dir, "non-a-dir")
		err := os.WriteFile(filePath, []byte{}, 0o600)
		must.NoError(err)

		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: filepath.Join(filePath, "logs"),
		})
		must.Error(err)
		must.Nil(log)
		must.ErrorContains(err, "mkdir log folder")
	})

	t.Run("open log file in read-only folder returns error", func(t *testing.T) {
		must := require.New(t)
		dir := t.TempDir()

		err := os.Chmod(dir, 0o500) //nolint:gosec // read-only intentionally to trigger OpenFile EACCES
		must.NoError(err)

		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: dir,
		})
		must.Error(err)
		must.Nil(log)
		must.ErrorContains(err, "open log file")
	})
}

//nolint:paralleltest // NewLogger writes to process-wide os.Stdout (hardcoded)
func TestLogger_Close(t *testing.T) {
	t.Run("close returns no error after successful init", func(t *testing.T) {
		must := require.New(t)
		dir := t.TempDir()

		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: dir,
		})
		must.NoError(err)
		must.NotNil(log)

		err = log.Close()
		must.NoError(err)
	})

	t.Run("double close returns error on second call", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)
		dir := t.TempDir()

		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: dir,
		})
		must.NoError(err)
		must.NotNil(log)

		err = log.Close()
		must.NoError(err)
		err = log.Close()
		must.Error(err)
		is.ErrorContains(err, "close log file")
	})
}

//nolint:paralleltest // NewLogger writes to process-wide os.Stdout (hardcoded)
func TestLogger_With(t *testing.T) {
	t.Run("returns new logger instance with field", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)
		dir := t.TempDir()

		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: dir,
		})
		must.NoError(err)
		must.NotNil(log)

		defer func() {
			is.NoError(log.Close(), "logger close should not fail")
		}()

		got := log.With(zap.String("request_id", "abc-123"))
		must.NotNil(got)
		is.NotSame(log, got, "With should return a new logger instance")
	})
}

//nolint:paralleltest // NewLogger writes to process-wide os.Stdout (hardcoded)
func TestLogger_ToContext(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	dir := t.TempDir()
	log, err := logger.NewLogger(logger.Config{
		Level:  "INFO",
		Folder: dir,
	})
	must.NoError(err)
	must.NotNil(log)

	defer func() {
		is.NoError(log.Close(), "logger close should not fail")
	}()

	ctx := logger.ToContext(context.Background(), log)

	got := logger.FromContext(ctx)
	is.Equal(log, got, "FromContext should return the same logger instance")
}

//nolint:paralleltest // NewLogger writes to process-wide os.Stdout (hardcoded)
func TestLogger_FromContext(t *testing.T) {
	t.Run("returns logger when present in context", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		dir := t.TempDir()
		log, err := logger.NewLogger(logger.Config{
			Level:  "INFO",
			Folder: dir,
		})
		must.NoError(err)

		defer func() {
			is.NoError(log.Close(), "logger close should not fail")
		}()

		ctx := logger.ToContext(context.Background(), log)

		got := logger.FromContext(ctx)
		must.NotNil(got)
		is.Equal(log, got, "FromContext should return the same logger instance")
	})

	t.Run("panics when no logger in context", func(t *testing.T) {
		is := assert.New(t)

		ctx := context.Background()

		is.PanicsWithValue("no log in context", func() {
			logger.FromContext(ctx)
		})
	})
}
