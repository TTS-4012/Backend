package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
)

type ContestsUsersRepoImp struct {
	conn *pgxpool.Pool
}

func (c *ContestsUsersRepoImp) Migrate(ctx context.Context) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS contests_users (
		contest_id int NOT NULL,
		user_id int NOT NULL,
		PRIMARY KEY (contest_id, user_id),
		FOREIGN KEY(contest_id) REFERENCES contests(id) ON DELETE CASCADE,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)
	`

	_, err := c.conn.Exec(ctx, stmt)
	return err
}

func NewContestsUsersRepo(ctx context.Context, conn *pgxpool.Pool) (db.ContestsUsersRepo, error) {
	ans := &ContestsUsersRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (c *ContestsUsersRepoImp) Add(ctx context.Context, contestID, userID int64) error {
	insertContestProblemsStmt := `
		INSERT INTO contests_users(
			contest_id, user_id) 
		VALUES($1, $2)
	`

	_, err := c.conn.Exec(ctx, insertContestProblemsStmt, contestID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *ContestsUsersRepoImp) Delete(ctx context.Context, contestID, userID int64) error {
	stmt := `
  DELETE FROM contests_users
  WHERE contest_id = $1 AND user_id = $2
  `
	_, err := c.conn.Exec(ctx, stmt, contestID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}
