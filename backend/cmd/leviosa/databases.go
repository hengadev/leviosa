package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GaryHY/leviosa/pkg/config"
	"github.com/GaryHY/leviosa/pkg/flags"
	"github.com/GaryHY/leviosa/pkg/redisutil"
	"github.com/GaryHY/leviosa/pkg/sqliteutil"

	"github.com/redis/go-redis/v9"
)

func setupDatabases(
	ctx context.Context,
	conf *config.Config,
	env mode.EnvMode,
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
		return nil, nil, fmt.Errorf("creating connection to redis database: %w", err)
	}

	sqlitedb, err := sqliteutil.Connect(ctx, sqliteutil.BuildDSN(sqliteConf.Filename))
	if err != nil {
		return nil, nil, fmt.Errorf("creating connection to sqlite database: %w", err)
	}

	// make new migration configuration
	migrationCfg, err := sqliteutil.NewMigrationConfig(sqlitedb, env)
	if err != nil {
		return nil, nil, fmt.Errorf("creating migration configuration: %w", err)
	}
	// run migration for database.
	if err := sqliteutil.SetMigrations(ctx, migrationCfg); err != nil {
		return nil, nil, fmt.Errorf("setting migration for SQLite database: %w", err)
	}

	// init databases
	queries, err := sqliteutil.GetInitQueries()
	if err != nil {
		return nil, nil, fmt.Errorf("getting init queries for SQLite database: %w", err)
	}
	if err := sqliteutil.Init(sqlitedb, queries...); err != nil {
		return sqlitedb, redisdb, fmt.Errorf("initialising SQLite database: %w", err)
	}

	return sqlitedb, redisdb, nil
}
