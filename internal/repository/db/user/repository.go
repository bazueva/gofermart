package user

import (
	"context"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"go.uber.org/zap"
)

type repository struct {
	db     interfaces.DB
	logger interfaces.Logger
}

const (
	defaultTimeout = 1 * time.Second
)

func (r *repository) CreateUser(ctx context.Context, user entities.User) (int32, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	query := table.Users.
		INSERT(table.Users.Login, table.Users.Password).
		VALUES(user.Login, user.Password).
		RETURNING(table.Users.ID.AS("id"))

	var result struct {
		ID int32
	}
	err := query.QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil {
		r.logger.Error("error repository CreateUser", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	return result.ID, nil
}

func (r *repository) ExistLogin(ctx context.Context, login string) (bool, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	query := postgres.SELECT(
		postgres.EXISTS(
			postgres.SELECT(
				table.Users.Login,
			).
				FROM(table.Users).
				WHERE(table.Users.Login.EQ(postgres.String(login))),
		).AS("exists"),
	)

	var response struct {
		Exists bool
	}
	err := query.QueryContext(ctxWithTimeout, r.db, &response)
	if err != nil {
		r.logger.Error("error ExistLogin", zap.Error(err))

		return false, entities.NewInternalServerError(err, "")
	}

	return response.Exists, nil
}

func NewRepository(db interfaces.DB, logger interfaces.Logger) *repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}
