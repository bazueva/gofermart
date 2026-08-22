package user

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRepository_ExistLogin(t *testing.T) {
	t.Parallel()

	t.Run("exist login", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).
				AddRow(true))

		exists, err := repo.ExistLogin(ctx, login)

		assert.Nil(t, err)
		assert.True(t, exists)
	})

	t.Run("not exist login", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).
				AddRow(false))

		exists, err := repo.ExistLogin(ctx, login)

		assert.Nil(t, err)
		assert.False(t, exists)
	})

	t.Run("error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("ошибка БД")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error ExistLogin", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnError(errorDB)

		exists, err := repo.ExistLogin(ctx, login)

		assert.Equal(t, "Internal Server Error", err.Error())
		assert.False(t, exists)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error ExistLogin", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnError(context.DeadlineExceeded)

		exists, err := repo.ExistLogin(ctx, login)

		assert.False(t, exists)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, err.(*entities.DomainError).ErrorType, entities.InternalServerErrorType)
		assert.Equal(t, "jet: context canceled", err.(*entities.DomainError).SourceErr.Error())
	})
}

func TestRepository_CreateUser(t *testing.T) {
	t.Run("success create user", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		user := entities.User{
			Login:    "testuser",
			Password: "password",
		}

		expectedSQL := `INSERT INTO public.users (login, password) VALUES ($1, $2) RETURNING users.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(user.Login, user.Password).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(100))

		userID, err := repo.CreateUser(ctx, user)

		assert.Nil(t, err)
		assert.Equal(t, int32(100), userID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error repository", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository CreateUser", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		user := entities.User{
			Login:    "testuser",
			Password: "password",
		}

		expectedSQL := `INSERT INTO public.users (login, password) VALUES ($1, $2) RETURNING users.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(user.Login, user.Password).
			WillReturnError(errors.New("connection refused"))

		userID, err := repo.CreateUser(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, err.(*entities.DomainError).ErrorType, entities.InternalServerErrorType)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository CreateUser", mock2.Anything)

		repo := NewRepository(db, logger)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		user := entities.User{
			Login:    "testuser",
			Password: "hashed_password",
		}

		expectedSQL := `INSERT INTO public.users (login, password) VALUES ($1, $2);`

		mock.ExpectQuery(expectedSQL).
			WithArgs(user.Login, user.Password).
			WillReturnError(context.DeadlineExceeded)

		userID, err := repo.CreateUser(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, err.(*entities.DomainError).ErrorType, entities.InternalServerErrorType)
		assert.Equal(t, "jet: context canceled", err.(*entities.DomainError).SourceErr.Error())
	})
}
