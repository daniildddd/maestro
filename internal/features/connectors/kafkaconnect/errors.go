package kafkaconnect

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("unexpected status %d: %s", e.StatusCode, e.Body)
}

func statusOf(err error) (int, bool) {
	var herr *HTTPError

	if !errors.As(err, &herr) {
		return 0, false
	}

	return herr.StatusCode, true
}

func bodyOf(err error) string {
	var herr *HTTPError

	if !errors.As(err, &herr) {
		return ""
	}

	return herr.Body
}

func isNotFound(err error) bool {
	status, ok := statusOf(err)

	return ok && status == http.StatusNotFound
}

func isAlreadyExists(err error) bool {
	status, ok := statusOf(err)

	return ok && status == http.StatusConflict && strings.Contains(bodyOf(err), "already exists")
}

func isRebalance(err error) bool {
	// Kafka Connect reports rebalance-in-progress as 409 with various messages
	// ("rebalance expected", "leader not known", ...); only 409 "already exists"
	// (duplicate create) is permanent, everything else on 409 is worth a retry.
	status, ok := statusOf(err)

	return ok && status == http.StatusConflict && !isAlreadyExists(err)
}

func isInvalidConfig(err error) bool {
	status, ok := statusOf(err)

	return ok && status == http.StatusBadRequest
}

func connectorError(err error) error {
	switch {
	case isNotFound(err):
		return domain.ErrConnectorNotFound
	case isAlreadyExists(err):
		return domain.ErrConnectorAlreadyExists
	case isRebalance(err):
		return domain.ErrRebalanceInProgress
	case isInvalidConfig(err):
		return domain.ErrInvalidConnectorConfig
	default:
		return err
	}
}
