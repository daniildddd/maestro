package cleanup_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/mock"

	"github.com/daniildddd/maestro/internal/core/logger"
	"github.com/daniildddd/maestro/internal/features/auth/cleanup"
)

type runResult struct {
	cancel context.CancelFunc

	deleteResults chan deleteResult
	done          chan struct{}
}

type deleteResult struct {
	deleted int64
	err     error
}

func runService(t *testing.T, interval time.Duration) runResult {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	repo := NewMockExpiredTokenCleaner(t)

	deleteResults := make(chan deleteResult, 1)

	repo.EXPECT().
		DeleteExpiredRefreshTokens(mock.Anything).
		RunAndReturn(func(_ context.Context) (int64, error) {
			result := <-deleteResults

			return result.deleted, result.err
		}).
		Maybe()

	svc := cleanup.NewWorker(
		repo,
		nopLogger(),
		cleanup.Config{Interval: interval},
	)

	done := make(chan struct{})

	go func() {
		defer close(done)

		svc.Run(ctx)
	}()

	return runResult{
		cancel:        cancel,
		deleteResults: deleteResults,
		done:          done,
	}
}

func waitForDelete(t *testing.T, results chan deleteResult) {
	t.Helper()

	select {
	case <-results:
	case <-time.After(5 * time.Second):
		t.Fatal("delete was not called before timeout")
	}
}

func TestWorker_RunCleansUpOnEveryTick(t *testing.T) {
	t.Parallel()

	rr := runService(t, 10*time.Millisecond)

	rr.deleteResults <- deleteResult{deleted: 5, err: nil}

	waitForDelete(t, rr.deleteResults)

	rr.deleteResults <- deleteResult{deleted: 2, err: nil}

	waitForDelete(t, rr.deleteResults)

	rr.cancel()

	select {
	case <-rr.done:
	case <-time.After(5 * time.Second):
		t.Fatal("service did not stop after cancel")
	}
}

func TestWorker_RunContinuesAfterError(t *testing.T) {
	t.Parallel()

	rr := runService(t, 10*time.Millisecond)

	rr.deleteResults <- deleteResult{deleted: 0, err: errors.New("boom")}

	waitForDelete(t, rr.deleteResults)

	rr.deleteResults <- deleteResult{deleted: 1, err: nil}

	waitForDelete(t, rr.deleteResults)

	rr.cancel()

	select {
	case <-rr.done:
	case <-time.After(5 * time.Second):
		t.Fatal("service did not stop after cancel")
	}
}

func TestWorker_RunStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	rr := runService(t, time.Hour)

	rr.cancel()

	select {
	case <-rr.done:
	case <-time.After(5 * time.Second):
		t.Fatal("service did not stop after cancel")
	}
}

func TestWorker_RunDoesNotTickBeforeInterval(t *testing.T) {
	t.Parallel()

	rr := runService(t, time.Hour)
	defer rr.cancel()

	select {
	case <-rr.deleteResults:
		t.Fatal("delete was called before any tick")
	case <-time.After(100 * time.Millisecond):
	}
}

func nopLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}
