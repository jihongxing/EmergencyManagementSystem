package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"emergency-management/backend/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Open(url string) (*sql.DB, error) {
	if url == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, errors.New("DATABASE_URL is invalid")
	}
	config.ConnectTimeout = 2 * time.Second
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

func Ready(db *sql.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := db.PingContext(ctx); err != nil {
			return errors.New("database unavailable")
		}
		var version int64
		var applied bool
		err := db.QueryRowContext(ctx, `SELECT version_id, is_applied FROM public.goose_db_version ORDER BY id DESC LIMIT 1`).Scan(&version, &applied)
		if err != nil || version != 1 || !applied {
			return errors.New("migration baseline unavailable")
		}
		return nil
	}
}

func Migrate(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.Files)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}
