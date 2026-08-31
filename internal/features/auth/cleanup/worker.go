package cleanup

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/daniildddd/maestro/internal/core/logger"
)

type ExpiredTokenCleaner interface {
	DeleteExpiredRefreshTokens(
		ctx context.Context,
	) (int64, error)
}

type Worker struct {
	repo     ExpiredTokenCleaner
	logger   *logger.Logger
	interval time.Duration
}

func NewWorker(
	repo ExpiredTokenCleaner,
	log *logger.Logger,
	config Config,
) *Worker {
	return &Worker{
		repo:     repo,
		logger:   log,
		interval: config.Interval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *Worker) runOnce(ctx context.Context) {
	deleted, err := w.repo.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		w.logger.Error("cleanup expired refresh tokens", zap.Error(err))

		return
	}

	if deleted > 0 {
		w.logger.Info("cleaned up expired refresh tokens", zap.Int64("deleted", deleted))
	}
}
