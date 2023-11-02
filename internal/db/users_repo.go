package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo interface {
	InsertUser(ctx context.Context, username, password, email string) (int, error)
}
type AuthRepoImp struct {
	conn *pgxpool.Pool
}

func NewAuthRepo(ctx context.Context, conn *pgxpool.Pool) (AuthRepo, error) {
	ans := &AuthRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *AuthRepoImp) Migrate(ctx context.Context) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS users(
	    id SERIAL PRIMARY KEY ,
	    username VARCHAR(40),
	    password varchar(70),
	    email varchar(40),
	    created_at TIMESTAMP DEFAULT NOW(),
	    is_valid boolean DEFAULT false,
	    UNIQUE (username),
	    UNIQUE (email)
	)
	`

	_, err := a.conn.Exec(ctx, stmt)
	return err
}
func (a *AuthRepoImp) InsertUser(ctx context.Context, username, email, password string) (int, error) {
	stmt := `
	INSERT INTO users(username, password, email) VALUES($1, $2, $3) RETURNING id
	`
	var id int
	err := a.conn.QueryRow(ctx, stmt, username, password, email).Scan(&id)
	return id, err
}
