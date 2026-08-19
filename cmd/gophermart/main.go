package main

import (
	"database/sql"
	"log"

	dbpkg "github.com/bazueva/gofermart/db"
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

}
