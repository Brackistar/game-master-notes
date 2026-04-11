package repos

import (
	"errors"
	"testing"

	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapFunctionError(t *testing.T) {
	notFoundTokens := map[string]struct{}{repoerrors.GMNNoteNotFound: {}}
	conflictTokens := map[string]struct{}{repoerrors.GMNNoteTagAlreadyOpen: {}}

	tests := []struct {
		name      string
		err       error
		assertErr error
	}{
		{
			name:      "pgx no rows maps to not found",
			err:       pgx.ErrNoRows,
			assertErr: repoerrors.ErrNotFound,
		},
		{
			name: "p0001 not found token maps to not found",
			err: &pgconn.PgError{
				Code:    "P0001",
				Message: repoerrors.GMNNoteNotFound,
			},
			assertErr: repoerrors.ErrNotFound,
		},
		{
			name: "p0001 conflict token maps to conflict",
			err: &pgconn.PgError{
				Code:    "P0001",
				Message: repoerrors.GMNNoteTagAlreadyOpen,
			},
			assertErr: repoerrors.ErrConflict,
		},
		{
			name: "p0001 unknown token maps to validation",
			err: &pgconn.PgError{
				Code:    "P0001",
				Message: "GMN_UNKNOWN_TOKEN",
			},
			assertErr: repoerrors.ErrValidation,
		},
		{
			name: "unique violation maps to conflict",
			err: &pgconn.PgError{
				Code: "23505",
			},
			assertErr: repoerrors.ErrConflict,
		},
		{
			name: "foreign key violation maps to validation",
			err: &pgconn.PgError{
				Code: "23503",
			},
			assertErr: repoerrors.ErrValidation,
		},
		{
			name:      "non postgres error maps to unknown",
			err:       errors.New("boom"),
			assertErr: repoerrors.ErrUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := mapFunctionError(tc.err, "test.op", "test_entity", notFoundTokens, conflictTokens)
			if !errors.Is(err, tc.assertErr) {
				t.Fatalf("expected error wrapping %v, got %v", tc.assertErr, err)
			}
		})
	}
}
