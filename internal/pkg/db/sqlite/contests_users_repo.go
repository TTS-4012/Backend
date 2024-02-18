package sqlite

import (
	"context"
	"database/sql"

	"github.com/ocontest/backend/internal/pkg/db/repos"
	"github.com/ocontest/backend/pkg/structs"

	"github.com/jackc/pgx/v5"
	"github.com/ocontest/backend/pkg"
	"github.com/pkg/errors"
)

type ContestsUsersRepoImp struct {
	conn *sql.DB
}

func (c *ContestsUsersRepoImp) Migrate(ctx context.Context) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS contests_users (
		contest_id int NOT NULL,
		user_id int NOT NULL,
		score float DEFAULT 0,
		PRIMARY KEY (contest_id, user_id),
		FOREIGN KEY(contest_id) REFERENCES contests(id) ON DELETE CASCADE,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)
	`

	_, err := c.conn.ExecContext(ctx, stmt)
	return err
}

func NewContestsUsersRepo(ctx context.Context, conn *sql.DB) (repos.ContestsUsersRepo, error) {
	ans := &ContestsUsersRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (c *ContestsUsersRepoImp) Add(ctx context.Context, contestID, userID int64) error {
	insertContestProblemsStmt := `
		INSERT INTO contests_users(
			contest_id, user_id) 
		VALUES($, $)
	`

	_, err := c.conn.ExecContext(ctx, insertContestProblemsStmt, contestID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *ContestsUsersRepoImp) Delete(ctx context.Context, contestID, userID int64) error {
	stmt := `
  DELETE FROM contests_users WHERE contest_id = $ AND user_id = $
  `
	_, err := c.conn.ExecContext(ctx, stmt, contestID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}

func (c *ContestsUsersRepoImp) IsRegistered(ctx context.Context, contestID, userID int64) (bool, error) {
	stmt := `
	SELECT EXISTS (SELECT 1 FROM contests_users WHERE contest_id = $ AND user_id = $)	
	`

	var exists bool
	err := c.conn.QueryRowContext(ctx, stmt, contestID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
func (c *ContestsUsersRepoImp) ListUsersByScore(ctx context.Context, contestID int64, limit, offset int) ([]structs.User, error) {
	args := make([]interface{}, 0)
	args = append(args, contestID)

	stmt := `
  	SELECT user_id, users.username FROM contests_users JOIN users ON contests_users.user_id = users.id WHERE contest_id = $1 ORDER BY score
  `

	if limit != 0 {
		args = append(args, limit)
		stmt += " LIMIT $"
	}
	if offset != 0 {
		args = append(args, offset)
		stmt += " OFFSET $"
	}

	rows, err := c.conn.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.Wrap(err, "coudn't run query stmt")
	}

	users := make([]structs.User, 0)
	for rows.Next() {
		var user structs.User

		err = rows.Scan(&user.ID, &user.Username)
		if err != nil {
			return users, errors.Wrap(err, "error on scan")
		}

		users = append(users, user)
	}
	return users, nil
}

func (c *ContestsUsersRepoImp) GetContestUsersCount(ctx context.Context, contestID int64) (int, error) {
	stmt := `
  	SELECT count(*) FROM contests_users WHERE contest_id = $
  	`

	var ans int
	err := c.conn.QueryRowContext(ctx, stmt, contestID).Scan(&ans)
	if err != nil {
		return 0, errors.Wrap(err, "coudn't run query stmt")
	}
	return ans, nil
}

// AddUserScore will add delta to current score of user.
func (c *ContestsUsersRepoImp) AddUserScore(ctx context.Context, userID, contestID int64, delta int) error {
	stmt := `
  		UPDATE contests_users SET score = score + $ WHERE contest_id = $ AND user_id = $
  	`
	_, err := c.conn.ExecContext(ctx, stmt, delta, contestID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}
