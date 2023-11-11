package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"ocontest/pkg/structs"
)

type AuthRepo interface {
	InsertUser(ctx context.Context, user structs.User) (int64, error)
	VerifyUser(ctx context.Context, userID int64) error
	GetByUsername(ctx context.Context, username string) (structs.User, error)
	GetByID(ctx context.Context, userID int64) (structs.User, error)
	UpdateUser(ctx context.Context, user structs.User) error
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
	    is_verified boolean DEFAULT false,
	    UNIQUE (username),
	    UNIQUE (email)
	)
	`

	_, err := a.conn.Exec(ctx, stmt)
	return err
}
func (a *AuthRepoImp) InsertUser(ctx context.Context, user structs.User) (int64, error) {
	stmt := `
	INSERT INTO users(username, password, email) VALUES($1, $2, $3) RETURNING id 
	`
	var id int64
	err := a.conn.QueryRow(ctx, stmt, user.Username, user.EncryptedPassword, user.Email).Scan(&id)
	return id, err
}

func (a *AuthRepoImp) VerifyUser(ctx context.Context, userID int64) error {
	stmt := `
	UPDATE users SET is_verified = true WHERE id = $1
	`
	_, err := a.conn.Exec(ctx, stmt, userID)
	return err
}

func (a *AuthRepoImp) GetByUsername(ctx context.Context, username string) (structs.User, error) {
	stmt := `
	SELECT id, username, password, email, is_verified FROM users WHERE username = $1 
	`
	var user structs.User
	err := a.conn.QueryRow(ctx, stmt, username).Scan(&user.ID, &user.Username, &user.EncryptedPassword, &user.Email, &user.Verified)
	return user, err
}

func (a *AuthRepoImp) GetByID(ctx context.Context, userID int64) (structs.User, error) {
	stmt := `
	SELECT id, username, password, email, is_verified FROM users WHERE id = $1 
	`
	var user structs.User
	err := a.conn.QueryRow(ctx, stmt, userID).Scan(&user.ID, &user.Username, &user.EncryptedPassword, &user.Email, &user.Verified)
	return user, err
}

// TODO: find a suitable query builder to do this shit. sorry for this shitty code you are gonna see, I had no other idea.
func (a *AuthRepoImp) UpdateUser(ctx context.Context, user structs.User) error {
	args := make([]interface{}, 0)
	args = append(args, user.ID)

	stmt := `
	UPDATE users SET
	`

	if user.Username != "" {
		args = append(args, user.Username)
		stmt += "username = $" + string(len(args))
	}
	if user.Email != "" {
		args = append(args, user.Email)
		if len(args) > 1 {
			stmt += ","
		}
		stmt += "email = $" + string(len(args))
	}
	if user.EncryptedPassword != "" {
		args = append(args, user.EncryptedPassword)
		stmt += "password = $" + string(len(args))
	}
	stmt += " WHERE id = $1"
	if len(args) == 0 {
		return nil
	}

	_, err := a.conn.Exec(ctx, stmt, args...)
	return err
}
