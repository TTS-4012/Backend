package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ocontest/backend/pkg/configs"
)

func NewConnectionPool(ctx context.Context, conf configs.SectionPostgres) (*pgxpool.Pool, error) {

	pgxConfig, _ := pgxpool.ParseConfig("")
	pgxConfig.ConnConfig.Host = conf.Host
	pgxConfig.ConnConfig.Port = uint16(conf.Port)
	pgxConfig.ConnConfig.Database = conf.Database
	pgxConfig.ConnConfig.User = conf.Username
	pgxConfig.ConnConfig.Password = conf.Password

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	return pool, err
}
