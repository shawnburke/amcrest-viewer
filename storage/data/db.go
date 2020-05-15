package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	"go.uber.org/config"
	"go.uber.org/fx"

	migrate "github.com/golang-migrate/migrate/v4"
	migratesql "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DBConfig struct {
	Database string
	DSN      string
	File     string
	Schemas  string
}

func NewDB(cfg config.Provider, lifecycle fx.Lifecycle) (*sqlx.DB, error) {

	dbCfg := &DBConfig{
		Database: "sqlite3",
		DSN:      ":memory:",
	}

	err := cfg.Get("database").Populate(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("Error getting DB config: %v", err)
	}

	db, err := sqlx.Connect(dbCfg.Database, dbCfg.DSN)

	if err != nil {
		return nil, err
	}

	lifecycle.Append(struct {
		OnStart func(context.Context) error
		OnStop  func(context.Context) error
	}{
		OnStart: func(ctx context.Context) error {
			return initDB(*dbCfg, db)
		},
		OnStop: func(ctx context.Context) error {
			if db != nil {
				return db.Close()
			}
			return nil
		},
	},
	)
	return db, err
}

func initDB(cfg DBConfig, db *sqlx.DB) error {

	// for some reason specifying this in the migration script doesn't
	// stick
	_, err := db.Exec("PRAGMA foreign_keys=on")
	if err != nil {
		return fmt.Errorf("Failed to set foreign_key pragma: %v", err)
	}

	driver, err := migratesql.WithInstance(db.DB, &migratesql.Config{})

	if _, err = os.Stat(cfg.Schemas); err != nil {
		pwd, _ := os.Getwd()
		fullDir := path.Join(pwd, cfg.Schemas)
		return fmt.Errorf("Can't find schemas: %v (dir=%s)", err, fullDir)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+cfg.Schemas,
		"main",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}

	return nil
}
