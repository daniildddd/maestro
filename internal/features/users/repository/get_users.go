package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func (r *UsersRepository) GetUsers(
	ctx context.Context,
	filter domain.UserFilter,
) ([]domain.User, error) {
	const op = "users.repository.GetUsers"

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query, args := buildGetUsersQuery(filter)

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

	var users []domain.User

	for rows.Next() {
		var dbUser userModel

		if err := dbUser.Scan(rows); err != nil {
			return nil, fmt.Errorf(
				"%s: scan row: %w",
				op,
				err,
			)
		}

		users = append(users, dbUser.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"%s: iterate rows: %w",
			op,
			err,
		)
	}

	return users, nil
}

func buildGetUsersQuery(filter domain.UserFilter) (query string, args []any) {
	var sb strings.Builder

	sb.WriteString(`
	SELECT id, username, password_hash, role, created_at, updated_at
	FROM users`)

	var conditions []string

	if filter.Username != "" {
		args = append(args, filter.Username)
		conditions = append(conditions, fmt.Sprintf("username=$%d", len(args)))
	}

	if filter.UserRole != "" {
		args = append(args, filter.UserRole)
		conditions = append(conditions, fmt.Sprintf("role=$%d", len(args)))
	}

	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	args = append(args, filter.Limit, offset(filter.Page, filter.Limit))
	fmt.Fprintf( //nolint:unhandled-error // writing to strings.Builder cannot fail
		&sb,
		" ORDER BY created_at LIMIT $%d OFFSET $%d",
		len(args)-1,
		len(args),
	)

	return sb.String(), args
}

func offset(page, limit int) int {
	return (page - 1) * limit
}
