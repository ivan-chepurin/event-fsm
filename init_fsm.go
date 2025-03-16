package event_fsm

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func initDB[T comparable](cfg *Config[T]) (*sqlx.DB, error) {
	dbConn, err := createDBConn(cfg)
	if err != nil {
		return nil, fmt.Errorf("createDBConn failed: %w", err)
	}

	if err = createSchema(cfg.DBConf, cfg.AppLabel); err != nil {
		return nil, fmt.Errorf("createSchema failed: %w", err)
	}

	cfg.DBConf = fmt.Sprintf("%s&search_path=%s", cfg.DBConf, searchPath)

	if err = dbConn.Close(); err != nil {
		return nil, fmt.Errorf("dbConn.Close failed: %w", err)
	}

	dbConn, err = createDBConn(cfg)
	if err != nil {
		return nil, fmt.Errorf("createDBConn failed: %w", err)
	}

	if err = Migrate(dbConn); err != nil {
		return nil, fmt.Errorf("Migrate failed: %w", err)
	}

	return dbConn, nil
}

func createSchema(dbConfig, appLabel string) error {
	connConfig, err := pgx.ParseConfig(dbConfig)
	if err != nil {
		return err
	}

	connConfig.RuntimeParams["application_name"] = appLabel
	connStr := stdlib.RegisterConnConfig(connConfig)

	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return err
	}

	if _, err = db.Exec(
		fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", searchPath),
	); err != nil {
		return err
	}

	if _, err = db.Exec(
		fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" SCHEMA %s", searchPath),
	); err != nil {
		return err
	}

	return nil
}

func createDBConn[T comparable](cfg *Config[T]) (*sqlx.DB, error) {

	connConfig, err := pgx.ParseConfig(cfg.DBConf)
	if err != nil {
		return nil, err
	}

	connConfig.RuntimeParams["application_name"] = cfg.AppLabel
	connStr := stdlib.RegisterConnConfig(connConfig)

	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionMaxLifetime)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis[T comparable](cfg *Config[T]) (*rClient, error) {
	var connectTimeLimit = time.Second * 5

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeLimit)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConf.URL,
		Password: cfg.RedisConf.Password,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis.Ping failed: %w", err)
	}

	return newRClient(client), nil
}
