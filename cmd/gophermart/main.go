package main

import (
	"database/sql"
	"log"
	"net/http"

	dbpkg "github.com/bazueva/gofermart/db"
	"github.com/bazueva/gofermart/internal/app"
	handlerPkg "github.com/bazueva/gofermart/internal/handler"
	"github.com/bazueva/gofermart/internal/middleware"
	"github.com/bazueva/gofermart/internal/repository/db/user"
	"github.com/bazueva/gofermart/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}

	cfg.logger, err = zap.NewProduction()
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
	userRepository := user.NewRepository(db, cfg.logger)

	userService := service.NewUserService(userRepository, cfg.logger, cfg.SecretKey)

	var application = app.NewApp(userService)
	handler := handlerPkg.NewHandler(cfg.logger, application)

	router := chi.NewRouter()
	router.Use(middleware.ServerLogger(cfg.logger))

	router.Post("/api/user/register", handler.RegisterUser)

	server := &http.Server{
		Addr:    cfg.ServerAddr.String(),
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		cfg.logger.Error("Ошибка сервера", zap.Error(err))
	}
}
