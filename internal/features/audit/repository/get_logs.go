package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (r *AuditRepository) GetLogs(
	ctx context.Context,
	filter *domain.AuditLogFilter,
) ([]domain.AuditEvent, error) {
	const op = "audit.repository.GetLogs"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query, args := buildGetLogsQuery(filter)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"%s: query (filter=%+v): %w",
			op,
			filter,
			err,
		)
	}
	defer rows.Close()

	var events []domain.AuditEvent

	for rows.Next() {
		var dbEvent auditEventModel

		if err := dbEvent.Scan(rows); err != nil {
			return nil, fmt.Errorf(
				"%s: scan row: %w",
				op,
				err,
			)
		}

		event, err := dbEvent.toDomain()
		if err != nil {
			return nil, fmt.Errorf(
				"%s: map event: %w",
				op,
				err,
			)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"%s: iterate rows: %w",
			op,
			err,
		)
	}

	return events, nil
}

func buildGetLogsQuery(filter *domain.AuditLogFilter) (query string, args []any) {
	var sb strings.Builder

	sb.WriteString(`
	SELECT id, action, outcome, failure_reason, actor_id, actor_login,
	       subject_type, subject_id, subject_name, state_before, state_after,
	       request_id, ip, user_agent, created_at
	FROM audit_logs`)

	var conditions []string

	if filter.Action != "" {
		args = append(args, string(filter.Action))
		conditions = append(conditions, fmt.Sprintf("action=$%d", len(args)))
	}

	if filter.Actor != "" {
		args = append(args, filter.Actor)
		conditions = append(conditions, fmt.Sprintf("actor_login=$%d", len(args)))
	}

	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	args = append(args, filter.Limit+1, offset(filter.Page, filter.Limit))
	fmt.Fprintf( //nolint:revive // writing to strings.Builder cannot fail
		&sb,
		" ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		len(args)-1,
		len(args),
	)

	return sb.String(), args
}

func offset(page, limit int) int {
	return (page - 1) * limit
}
