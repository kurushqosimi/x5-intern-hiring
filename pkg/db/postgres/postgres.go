package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, dsn string) (pool *pgxpool.Pool, err error) {
	err = DoWithAttempts(func() error {
		dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var config *pgxpool.Config
		config, err = pgxpool.ParseConfig(dsn)
		if err != nil {
			return fmt.Errorf("pgxpool.ParseConfig: %w", err)
		}

		pool, err = pgxpool.NewWithConfig(dbCtx, config)
		if err != nil {
			return fmt.Errorf("pgxpool.NewWithConfig: %w", err)
		}

		if pingErr := pool.Ping(dbCtx); pingErr != nil {
			pool.Close()
			return fmt.Errorf("ping: %w", pingErr)
		}

		return nil
	}, 5, 5*time.Second)

	if err != nil {
		return nil, fmt.Errorf("all attempts are exceeded. Unable to connect to postgres: %w", err)
	}

	return pool, nil
}

func DoWithAttempts(fn func() error, maxAttempts int, delay time.Duration) error {
	var err error

	for maxAttempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			maxAttempts--

			continue
		}

		return nil
	}

	return err
}
