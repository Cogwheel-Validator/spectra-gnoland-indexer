package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TimescaleDb struct {
	pool *pgxpool.Pool
}

func NewTimescaleDb(config DatabasePoolConfig) *TimescaleDb {
	return &TimescaleDb{
		pool: nil,
	}
}

func (t *TimescaleDb) ConnectToDb(config DatabasePoolConfig) error {
	parseConfig, err := pgxpool.ParseConfig(
		fmt.Sprintf(
			`
			host=%s 
			port=%d 
			user=%s 
			password=%s 
			dbname=%s 
			sslmode=%s 
			pool_max_conns=%v 
			pool_min_conns=%v 
			pool_max_conn_lifetime=%s 
			pool_max_conn_idle_time=%s 
			pool_health_check_period=%s 
			pool_max_conn_lifetime_jitter=%s
			`,
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.Dbname,
			config.Sslmode,
			config.PoolMaxConns,
			config.PoolMinConns,
			config.PoolMaxConnLifetime,
			config.PoolMaxConnIdleTime,
			config.PoolHealthCheckPeriod,
			config.PoolMaxConnLifetimeJitter))
	if err != nil {
		return err
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return err
	}
	t.pool = conn
	return nil
}
