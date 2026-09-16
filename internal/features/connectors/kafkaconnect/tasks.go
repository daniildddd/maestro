package kafkaconnect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (c *HTTPClient) GetTaskByID(
	ctx context.Context,
	name string,
	taskID int,
) (domain.Task, error) {
	const op = "connectors.kafkaconnect.GetTaskByID"

	path := fmt.Sprintf("/connectors/%s/tasks/%d/status", url.PathEscape(name), taskID)

	var state struct {
		ID       int     `json:"id"`
		State    string  `json:"state"`
		WorkerID string  `json:"worker_id"`
		Trace    *string `json:"trace"`
	}

	if err := c.do(ctx, http.MethodGet, path, nil, &state); err != nil {
		if isNotFound(err) {
			return domain.Task{}, fmt.Errorf("%s: %w", op, domain.ErrConnectorTaskNotFound)
		}

		return domain.Task{}, fmt.Errorf("%s: %w", op, err)
	}

	return domain.Task{
		ID:       state.ID,
		State:    strings.ToLower(state.State),
		WorkerID: state.WorkerID,
		Trace:    state.Trace,
	}, nil
}

func (c *HTTPClient) RestartTask(ctx context.Context, name string, taskID int) error {
	const op = "connectors.kafkaconnect.RestartTask"

	path := fmt.Sprintf("/connectors/%s/tasks/%d/restart", url.PathEscape(name), taskID)

	err := c.withRetry(ctx, isRebalance, func() error {
		return c.do(ctx, http.MethodPost, path, nil, nil)
	})
	if err == nil {
		return nil
	}

	if isNotFound(err) {
		return fmt.Errorf("%s: %w", op, domain.ErrConnectorTaskNotFound)
	}

	return fmt.Errorf("%s: %w", op, err)
}
