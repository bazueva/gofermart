package helpers

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func SQLMockTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, error) {
	return sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
			cleanExpected := strings.TrimSuffix(strings.TrimSpace(strings.ToLower(expectedSQL)), ";")
			cleanActual := strings.TrimSuffix(strings.TrimSpace(strings.ToLower(actualSQL)), ";")

			cleanExpected = strings.Join(strings.Fields(cleanExpected), " ")
			cleanActual = strings.Join(strings.Fields(cleanActual), " ")

			if cleanExpected != cleanActual {
				t.Errorf("EXPECTED:\n%s\n\nACTUAL (из Jet):\n%s\n", expectedSQL, actualSQL)
				return fmt.Errorf("sql %q does not match %q", expectedSQL, actualSQL)
			}
			return nil
		})),
	)
}
