package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokensRepo interface {
	InsertUser(ctx context.Context, username, password, email string) (int, error)
}
type TokensRepoImp struct {
	conn *pgxpool.Pool
}

func NewTokensRepo(ctx context.Context, conn *pgxpool.Pool) (TokensRepo, error) {
	ans := &TokensRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *TokensRepoImp) Migrate(ctx context.Context) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS users(
	    id SERIAL PRIMARY KEY ,
	    username VARCHAR(40),
	    password varchar(70),
	    email varchar(40),
	    created_at TIMESTAMP DEFAULT NOW(),
	    UNIQUE (username),
	    UNIQUE (email)
	)
	`

	_, err := a.conn.Exec(ctx, stmt)
	return err
}
func (a *TokensRepoImp) InsertUser(ctx context.Context, username, email, password string) (int, error) {
	stmt := `
	INSERT INTO users(username, password, email) VALUES($1, $2, $3) RETURNING id
	`
	var id int
	err := a.conn.QueryRow(ctx, stmt, username, password, email).Scan(&id)
	return id, err
}
