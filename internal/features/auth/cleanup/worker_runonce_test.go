package cleanup

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/logger"
)

type stubCleaner struct {
	deleted int64
	err     error
	calls   int
}

func (s *stubCleaner) DeleteExpiredRefreshTokens(_ context.Context) (int64, error) {
	s.calls++

	return s.deleted, s.err
}

func newObservedWorker(
	t *testing.T,
	repo ExpiredTokenCleaner,
	lvl zapcore.Level,
) (*Worker, *observer.ObservedLogs) {
	t.Helper()

	core, recorded := observer.New(lvl)

	return NewWorker(
		repo,
		&logger.Logger{Logger: zap.New(core)},
		Config{Interval: time.Hour},
	), recorded
}

func TestWorker_runOnce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		deleted     int64
		err         error
		wantCalls   int
		wantLogs    int
		wantLevel   zapcore.Level
		wantMessage string
		wantDeleted int64
	}{
		{
			name:        "error is logged and swallowed",
			deleted:     0,
			err:         errors.New("boom"),
			wantCalls:   1,
			wantLogs:    1,
			wantLevel:   zapcore.ErrorLevel,
			wantMessage: "cleanup expired refresh tokens",
		},
		{
			name:        "deleted count is logged when positive",
			deleted:     5,
			err:         nil,
			wantCalls:   1,
			wantLogs:    1,
			wantLevel:   zapcore.InfoLevel,
			wantMessage: "cleaned up expired refresh tokens",
			wantDeleted: 5,
		},
		{
			name:      "no log when nothing deleted",
			deleted:   0,
			err:       nil,
			wantCalls: 1,
			wantLogs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			repo := &stubCleaner{deleted: tt.deleted, err: tt.err}

			worker, recorded := newObservedWorker(t, repo, zapcore.DebugLevel)

			worker.runOnce(context.Background())

			must.Equal(tt.wantCalls, repo.calls)
			must.Len(recorded.All(), tt.wantLogs)

			if tt.wantLogs == 0 {
				return
			}

			logs := recorded.All()
			must.Equal(tt.wantLevel, logs[0].Level)
			must.Equal(tt.wantMessage, logs[0].Message)

			if tt.wantDeleted > 0 {
				must.Equal(tt.wantDeleted, logs[0].ContextMap()["deleted"])
			}
		})
	}
}
