package data

import (
	"context"
	"fmt"
	"log"

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
}

const memoryDSN = ":memory:"

func NewConfig(cfg config.Provider) (*DBConfig, error) {
	dbCfg := &DBConfig{
		Database: "sqlite3",
	}

	err := cfg.Get("database").Populate(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("Error getting DB config: %v", err)
	}
	return dbCfg, nil
}

func New(dbCfg *DBConfig, lifecycle fx.Lifecycle) (*sqlx.DB, error) {

	if dbCfg.Database == "" {
		dbCfg.Database = "sqlite3"
	}

	if dbCfg.DSN == "" {
		dbCfg.DSN = memoryDSN
	}

	db, err := sqlx.Connect(dbCfg.Database, dbCfg.DSN)

	if err != nil {
		return nil, fmt.Errorf("Can't connect to DB (DSN=%s): %w", dbCfg.DSN, err)
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

	schemaPath, done, err := getSchemaDir()
	defer done()

	if err != nil {
		return fmt.Errorf("Failed to write schema files: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+schemaPath,
		"main",
		driver,
	)
	if err != nil {
		log.Panic(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	return nil
}
