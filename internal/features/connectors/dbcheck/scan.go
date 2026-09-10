package dbcheck

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func scanString(ctx context.Context, conn *pgx.Conn, query string, args ...any) (string, error) {
	var value string

	err := conn.QueryRow(ctx, query, args...).Scan(&value)

	return value, err
}

func scanInt(ctx context.Context, conn *pgx.Conn, query string, args ...any) (int, error) {
	var value int

	err := conn.QueryRow(ctx, query, args...).Scan(&value)

	return value, err
}

func scanBool(ctx context.Context, conn *pgx.Conn, query string, args ...any) (bool, error) {
	var value bool

	err := conn.QueryRow(ctx, query, args...).Scan(&value)

	return value, err
}
