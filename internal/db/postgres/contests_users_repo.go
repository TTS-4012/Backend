package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
	"github.com/pkg/errors"
)

type ContestsUsersRepoImp struct {
	conn *pgxpool.Pool
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
  DELETE FROM contests_users WHERE contest_id = $1 AND user_id = $2
  `
	_, err := c.conn.Exec(ctx, stmt, contestID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}

func (c *ContestsUsersRepoImp) ListUsersByScore(ctx context.Context, contestID int64, limit, offset int) ([]int64, error) {
	stmt := `
  	SELECT user_id FROM contests_users WHERE contest_id = $1 ORDER BY score LIMIT $2 OFFSET $3
  	`

	rows, err := c.conn.Query(ctx, stmt, contestID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "coudn't run query stmt")
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64

		err = rows.Scan(&id)
		if err != nil {
			return ids, errors.Wrap(err, "error on scan")
		}

		ids = append(ids, id)
	}
	return ids, nil
}

func (c *ContestsUsersRepoImp) GetContestUsersCount(ctx context.Context, contestID int64) (int, error) {
	stmt := `
  	SELECT count(*) FROM contests_users WHERE contest_id = $1 
  	`

	var ans int
	err := c.conn.QueryRow(ctx, stmt, contestID).Scan(&ans)
	if err != nil {
		return 0, errors.Wrap(err, "coudn't run query stmt")
	}
	return ans, nil
}

// AddUserScore will add delta to current score of user.
func (c *ContestsUsersRepoImp) AddUserScore(ctx context.Context, userID, contestID int64, delta int) error {
	stmt := `
  		UPDATE contests_users SET score = score + $1 WHERE contest_id = $2 AND user_id = $3
  	`
	_, err := c.conn.Exec(ctx, stmt, delta, contestID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err

}
