package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"
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
		created_at TIMESTAMP DEFAULT NOW(),
		CONSTRAINT fk_created_by_contest FOREIGN KEY(created_by) REFERENCES users(id)
	);
	
	CREATE TABLE IF NOT EXISTS contest_problems (
		contest_id int NOT NULL,
		problem_id int NOT NULL,
		CONSTRAINT pk_contest_problems PRIMARY KEY (contest_id, problem_id),
		CONSTRAINT fk_contest FOREIGN KEY(contest_id) REFERENCES contests(id),
		CONSTRAINT fk_problem FOREIGN KEY(problem_id) REFERENCES problems(id)
	)
	`

	_, err := c.conn.Exec(ctx, stmt)
	return err
}

func NewContestsMetadataRepo(ctx context.Context, conn *pgxpool.Pool) (db.ContestsMetadataRepo, error) {
	ans := &ContestsMetadataRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (c *ContestsMetadataRepoImp) InsertContest(ctx context.Context, contest structs.Contest) (int64, error) {
	insertContestStmt := `
		INSERT INTO contests(
			created_by, title) 
		VALUES($1, $2) RETURNING id
	`

	var contestID int64
	err := c.conn.QueryRow(ctx, insertContestStmt, contest.CreatedBy, contest.Title).Scan(&contestID)
	if err != nil {
		return 0, err
	}

	insertContestProblemsStmt := `
		INSERT INTO contest_problems(
			contest_id, problem_id) 
		VALUES($1, $2)
	`

	for _, problem := range contest.Problems {
		_, err := c.conn.Exec(ctx, insertContestProblemsStmt, contestID, problem.ID)
		if err != nil {
			return 0, err
		}
	}

	return contestID, nil
}

func (c *ContestsMetadataRepoImp) GetContest(ctx context.Context, id int64) (structs.Contest, error) {
	selectContestStmt := `
		SELECT created_by, title FROM contests WHERE id = $1
	`

	var contest structs.Contest
	err := c.conn.QueryRow(ctx, selectContestStmt, id).
		Scan(&contest.CreatedBy, &contest.Title)
	if errors.Is(err, pgx.ErrNoRows) {
		return structs.Contest{}, pkg.ErrNotFound
	} else if err != nil {
		return structs.Contest{}, err
	}

	selectContestProblemsStmt := `
		SELECT problem_id FROM contest_problems WHERE contest_id = $1
	`

	rows, err := c.conn.Query(ctx, selectContestProblemsStmt, id)
	if err != nil {
		return structs.Contest{}, err
	}
	defer rows.Close()

	problemsRespo, err := NewProblemsMetadataRepo(ctx, c.conn)
	if err != nil {
		return structs.Contest{}, err
	}

	for rows.Next() {
		var problemID int64
		err := rows.Scan(&problemID)
		if err != nil {
			return structs.Contest{}, err
		}

		problem, err := problemsRespo.GetProblem(ctx, problemID)
		if err != nil {
			return structs.Contest{}, err
		}

		contest.Problems = append(contest.Problems, problem)
	}

	return contest, nil
}

func (c *ContestsMetadataRepoImp) ListContests(ctx context.Context) ([]structs.Contest, error) {
	return nil, nil
}
