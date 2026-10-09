package repos

import (
	"errors"
	"strings"

	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapFunctionError(
	err error,
	op string,
	entity string,
	notFoundTokens map[string]struct{},
	conflictTokens map[string]struct{},
) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return repoerrors.NewNotFound(op, entity)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "P0001" {
			token := strings.TrimSpace(pgErr.Message)
			if _, ok := notFoundTokens[token]; ok {
				return repoerrors.NewNotFound(op, entity)
			}
			if _, ok := conflictTokens[token]; ok {
				return repoerrors.NewConflict(op, entity)
			}
			return repoerrors.WrapValidation(op, entity, err)
		}

		if pgErr.Code == "23505" {
			return repoerrors.NewConflict(op, entity)
		}
		if pgErr.Code == "23503" || pgErr.Code == "23514" || pgErr.Code == "22P02" {
			return repoerrors.WrapValidation(op, entity, err)
		}
	}

	return repoerrors.WrapUnknown(op, entity, err)
}
