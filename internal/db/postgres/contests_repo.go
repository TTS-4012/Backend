package postgres

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
)

type ContestsMetadataRepoImp struct {
	conn *pgxpool.Pool
}

func (c *ContestsMetadataRepoImp) Migrate(ctx context.Context) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS contests (
		id SERIAL PRIMARY KEY,
		created_by int NOT NULL,
		title varchar(70) NOT NULL,
	    start_time bigint NOT NULL,
	    duration int NOT NULL,
		created_at TIMESTAMP DEFAULT NOW(),
		CONSTRAINT fk_created_by_contest FOREIGN KEY(created_by) REFERENCES users(id)
	);
	`

	_, err := c.conn.Exec(ctx, stmt)
	return err
}

func NewContestsMetadataRepo(ctx context.Context, conn *pgxpool.Pool) (db.ContestsMetadataRepo, error) {
	ans := &ContestsMetadataRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (c *ContestsMetadataRepoImp) InsertContest(ctx context.Context, contest structs.Contest) (int64, error) {
	var contestID int64
	insertContestStmt := `
			INSERT INTO contests(
				created_by, title, start_time, duration) 
			VALUES($1, $2, $3, $4) RETURNING id
		`

	err := c.conn.
		QueryRow(ctx, insertContestStmt, contest.CreatedBy, contest.Title, contest.StartTime, contest.Duration).
		Scan(&contestID)
	if err != nil {
		return 0, err
	}

	return contestID, nil
}

func (c *ContestsMetadataRepoImp) GetContest(ctx context.Context, id int64) (structs.Contest, error) {
	selectContestStmt := `
		SELECT created_by, title, start_time, duration FROM contests WHERE id = $1
	`

	var contest structs.Contest
	err := c.conn.QueryRow(ctx, selectContestStmt, id).
		Scan(&contest.CreatedBy, &contest.Title, &contest.StartTime, &contest.Duration)
	if errors.Is(err, pgx.ErrNoRows) {
		return structs.Contest{}, pkg.ErrNotFound
	} else if err != nil {
		return structs.Contest{}, err
	}

	return contest, nil
}

func (c *ContestsMetadataRepoImp) ListContests(ctx context.Context, descending bool, limit, offset int, started bool, userID int64, owned, getCount bool) ([]structs.Contest, int, error) {
	stmt := `
	SELECT id, created_by, title, start_time, duration
	`
	if getCount {
		stmt = fmt.Sprintf("%s, COUNT(*) OVER() AS total_count", stmt)
	}

	stmt = fmt.Sprintf("%s FROM contests", stmt)

	now := time.Now().Unix()
	if started {
		stmt = fmt.Sprintf("%s WHERE start_time <= %d", stmt, now)
	} else {
		stmt = fmt.Sprintf("%s WHERE start_time > %d", stmt, now)
	}

	if owned {
		stmt = fmt.Sprintf("%s AND created_by = $1", stmt)
	}

	stmt += " ORDER BY id "

	if descending {
		stmt += " DESC"
	}
	if limit != 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, limit)
	}
	if offset != 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, offset)
	}

	var rows pgx.Rows
	var err error
	if owned {
		rows, err = c.conn.Query(ctx, stmt, userID)
	} else {
		rows, err = c.conn.Query(ctx, stmt)
	}
	if err != nil {
		return nil, 0, err
	}

	ans := make([]structs.Contest, 0)
	var total_count int = 0
	for rows.Next() {

		var contest structs.Contest
		if getCount {
			err = rows.Scan(
				&contest.ID,
				&contest.CreatedBy,
				&contest.Title,
				&contest.StartTime,
				&contest.Duration,
				&total_count,
			)
		} else {
			err = rows.Scan(
				&contest.ID,
				&contest.CreatedBy,
				&contest.Title,
				&contest.StartTime,
				&contest.Duration,
			)
		}
		if err != nil {
			return nil, 0, err
		}
		ans = append(ans, contest)
	}
	return ans, total_count, err
}

func (c *ContestsMetadataRepoImp) UpdateContests(ctx context.Context, id int64, newContest structs.RequestUpdateContest) error {
	existCheckingStmnt := `SELECT id FROM contests WHERE id = $1`
	if err := c.conn.QueryRow(ctx, existCheckingStmnt, id).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = pkg.ErrNotFound
		}
		return err
	}

	if newContest.Title != "" {

		stmt := `
			UPDATE contests SET title = $1
			WHERE id = $2
			RETURNING id;
		`
		_, err := c.conn.Exec(ctx, stmt, newContest.Title, id)
		if err != nil {
			return err
		}
	}

	if newContest.StartTime != 0 {
		stmt := `
		UPDATE contests SET start_time = $1 WHERE id = $2
		`
		_, err := c.conn.Exec(ctx, stmt, newContest.StartTime, id)
		if err != nil {
			return err
		}
	}

	if newContest.Duration != 0 {
		stmt := `
		UPDATE contests SET duration = $1 WHERE id = $2
		`
		_, err := c.conn.Exec(ctx, stmt, newContest.Duration, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ContestsMetadataRepoImp) DeleteContest(ctx context.Context, id int64) error {
	stmt := `
	DELETE FROM contests WHERE id = $1
	`
	_, err := c.conn.Exec(ctx, stmt, id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}

func (c *ContestsMetadataRepoImp) ListMyContests(ctx context.Context, descending bool, limit, offset int, started bool, userID int64, getCount bool) ([]structs.Contest, int, error) {
	stmt := `
	SELECT id, created_by, title, start_time, duration
	`

	if getCount {
		stmt = fmt.Sprintf("%s, COUNT(*) OVER() AS total_count", stmt)
	}

	stmt = fmt.Sprintf("%s FROM contests JOIN contests_users ON contests_users.contest_id = contests.id WHERE contests_users.user_id = $1", stmt)

	now := time.Now().Unix()
	if started {
		stmt = fmt.Sprintf("%s AND start_time <= %d", stmt, now)
	} else {
		stmt = fmt.Sprintf("%s AND start_time > %d", stmt, now)
	}

	stmt += " ORDER BY id "

	if descending {
		stmt += " DESC"
	}
	if limit != 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, limit)
	}
	if offset != 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, offset)
	}

	rows, err := c.conn.Query(ctx, stmt, userID)
	if err != nil {
		return nil, 0, err
	}

	ans := make([]structs.Contest, 0)
	var total_count int = 0
	for rows.Next() {

		var contest structs.Contest
		if getCount {
			err = rows.Scan(
				&contest.ID,
				&contest.CreatedBy,
				&contest.Title,
				&contest.StartTime,
				&contest.Duration,
				&total_count,
			)
		} else {
			err = rows.Scan(
				&contest.ID,
				&contest.CreatedBy,
				&contest.Title,
				&contest.StartTime,
				&contest.Duration,
			)
		}
		if err != nil {
			return nil, 0, err
		}
		ans = append(ans, contest)
	}
	return ans, total_count, err
}
