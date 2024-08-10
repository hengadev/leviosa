package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	//utils
	"github.com/GaryHY/event-reservation-app/pkg/config"
	"github.com/GaryHY/event-reservation-app/pkg/redisutil"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"

	// external packages
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func setupDatabases(
	ctx context.Context,
	conf *config.Config,
) (*sql.DB, *redis.Client, error) {
	sqliteConf := conf.GetSQLITE()
	redisConf := conf.GetRedis()

	// databases setup
	redisdb, err := redisutil.Connect(
		ctx,
		redisutil.WithAddr(redisConf.Addr),
		redisutil.WithDB(redisConf.DB),
		redisutil.WithPassword(redisConf.Password),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection to redis : %w", err)
	}

	sqlitedb, err := sqliteutil.Connect(ctx, sqliteutil.BuildDSN(sqliteConf.Filename))
	if err != nil {
		return nil, nil, fmt.Errorf("create connection to sqlite : %w", err)
	}

	// setup goose
	goose.SetBaseFS(nil)
	// Set the dialect to SQLite3
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, nil, fmt.Errorf("Failed to set dialect: %w", err)
	}
	// run the migration to the database.
	if err := goose.UpContext(ctx, sqlitedb, os.Getenv("MIGRATION_PATH")); err != nil {
		return nil, nil, fmt.Errorf("failed to run migration %w", err)
	}

	// init hte database
	queries, err := sqliteutil.GetInitQueries()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get init queries for sqlite database: %w", err)
	}
	if err := sqliteutil.Init(sqlitedb, queries...); err != nil {
		return sqlitedb, redisdb, fmt.Errorf("failed to init sqlite database: %w", err)
	}

	return sqlitedb, redisdb, nil
}
