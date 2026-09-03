package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	dbpkg "github.com/bazueva/gofermart/db"
	"github.com/bazueva/gofermart/internal/app"
	handlerPkg "github.com/bazueva/gofermart/internal/handler"
	"github.com/bazueva/gofermart/internal/middleware"
	"github.com/bazueva/gofermart/internal/repository/bonus"
	"github.com/bazueva/gofermart/internal/repository/db/order"
	"github.com/bazueva/gofermart/internal/repository/db/user"
	orderService "github.com/bazueva/gofermart/internal/service/order"
	userService "github.com/bazueva/gofermart/internal/service/user"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}

	cfg.logger, err = zap.NewProduction(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = cfg.logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if cfg.DatabaseDSN != "" {
		if err := dbpkg.RunMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
	}

	startServer(cfg, db)
}

func startServer(cfg config, db *sql.DB) {
	/* http repo */
	bonusRepository, err := bonus.NewRepository(
		cfg.AccrualSystemAddress,
		cfg.logger,
	)
	if err != nil {
		panic(err)
	}

	/* DB repo */
	userRepository := user.NewRepository(db, cfg.logger)
	orderRepository := order.NewRepository(db, cfg.logger)

	/* workers */
	ctx := context.Background()
	orderProcessor := orderService.NewOrderProcessor(bonusRepository, orderRepository, cfg.logger)
	orderProcessor.Start(ctx)
	orderProcessor.StartDatabasePoller(ctx)

	/* services */
	userService := userService.NewUserService(userRepository, cfg.logger, cfg.SecretKey)
	orderService := orderService.NewOrder(orderRepository, orderProcessor, cfg.logger)

	var application = app.NewApp(userService, orderService, cfg.logger)
	handler := handlerPkg.NewHandler(cfg.logger, application)

	router := chi.NewRouter()
	router.Use(middleware.ServerLogger(cfg.logger))
	router.Use(middleware.JSONMiddleware)

	router.Post("/api/user/register", handler.RegisterUser)
	router.Post("/api/user/login", handler.LoginUser)

	router.Group(func(r chi.Router) {
		r.Use(middleware.Authorization(application, cfg.logger))

		r.Post("/api/user/orders", handler.CreateOrder)
		r.Get("/api/user/orders", handler.UserOrdersList)
	})

	server := &http.Server{
		Addr:    cfg.ServerAddr.String(),
		Handler: router,
	}

	cfg.logger.Info("Сервер запущен по адресу", zap.String("addr", server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		cfg.logger.Error("Ошибка сервера", zap.Error(err))
	}
}
