package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation = "23505"
	pgFKViolation     = "23503"
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func IsUniqueViolation(err error) bool     { return pgCode(err) == pgUniqueViolation }
func IsForeignKeyViolation(err error) bool { return pgCode(err) == pgFKViolation }
