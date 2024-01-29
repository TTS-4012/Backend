package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ocontest/backend/internal/db/repos"

	"github.com/ocontest/backend/pkg/structs"
)

type UsersRepoImp struct {
	conn *sql.DB
}

func NewAuthRepo(ctx context.Context, conn *sql.DB) (repos.UsersRepo, error) {
	ans := &UsersRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *UsersRepoImp) Migrate(ctx context.Context) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS users(
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    username VARCHAR(40),
	    password varchar(70),
	    email varchar(40),
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    is_verified boolean DEFAULT false,
	    UNIQUE (username),
	    UNIQUE (email)
	)
	`

	_, err := a.conn.ExecContext(ctx, stmt)
	return err
}
func (a *UsersRepoImp) InsertUser(ctx context.Context, user structs.User) (int64, error) {
	stmt := `
	INSERT INTO users(username, password, email) VALUES($, $, $) RETURNING id 
	`
	var id int64
	err := a.conn.QueryRowContext(ctx, stmt, user.Username, user.EncryptedPassword, user.Email).Scan(&id)
	return id, err
}

func (a *UsersRepoImp) VerifyUser(ctx context.Context, userID int64) error {
	stmt := `
	UPDATE users SET is_verified = true WHERE id = $
	`
	_, err := a.conn.ExecContext(ctx, stmt, userID)
	return err
}

func (a *UsersRepoImp) GetByUsername(ctx context.Context, username string) (structs.User, error) {
	stmt := `
	SELECT id, username, password, email, is_verified FROM users WHERE username = $ 
	`
	var user structs.User
	err := a.conn.QueryRowContext(ctx, stmt, username).Scan(&user.ID, &user.Username, &user.EncryptedPassword, &user.Email, &user.Verified)
	return user, err
}

func (a *UsersRepoImp) GetByID(ctx context.Context, userID int64) (structs.User, error) {
	stmt := `
	SELECT id, username, password, email, is_verified FROM users WHERE id = $ 
	`
	var user structs.User
	err := a.conn.QueryRowContext(ctx, stmt, userID).Scan(&user.ID, &user.Username, &user.EncryptedPassword, &user.Email, &user.Verified)
	return user, err
}

func (a *UsersRepoImp) GetUsername(ctx context.Context, userID int64) (string, error) {
	stmt := `
	SELECT username FROM users WHERE id = $ 
	`
	var username string
	err := a.conn.QueryRowContext(ctx, stmt, userID).Scan(&username)
	return username, err
}

func (a *UsersRepoImp) GetByEmail(ctx context.Context, email string) (structs.User, error) {
	stmt := `
	SELECT id, username, password, email, is_verified FROM users WHERE email = $ 
	`
	var user structs.User
	err := a.conn.QueryRowContext(ctx, stmt, email).Scan(&user.ID, &user.Username, &user.EncryptedPassword, &user.Email, &user.Verified)
	return user, err
}

// TODO: find a suitable query builder to do this shit. sorry for this shitty code you are gonna see, I had no other idea.
func (a *UsersRepoImp) UpdateUser(ctx context.Context, user structs.User) error {
	args := make([]interface{}, 0)
	args = append(args, user.ID)

	stmt := `
	UPDATE users SET
	`

	if user.Username != "" {
		args = append(args, user.Username)
		stmt += fmt.Sprintf("username = $%d", len(args))
	}
	if user.Email != "" {
		args = append(args, user.Email)
		if len(args) > 1 {
			stmt += ","
		}
		stmt += fmt.Sprintf("email = $%d", len(args))
	}
	if user.EncryptedPassword != "" {
		args = append(args, user.EncryptedPassword)
		stmt += fmt.Sprintf("password = $%d", len(args))
	}
	stmt += " WHERE id = $"
	if len(args) == 0 {
		return nil
	}

	_, err := a.conn.ExecContext(ctx, stmt, args...)
	return err
}
