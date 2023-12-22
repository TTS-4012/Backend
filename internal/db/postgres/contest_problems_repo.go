package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ocontest/backend/internal/db"
)

type ContestsProblemsRepoImp struct {
	conn *pgxpool.Pool
}

func (c *ContestsProblemsRepoImp) Migrate(ctx context.Context) error {
	stmt := `
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

func NewContestsProblemsRepo(ctx context.Context, conn *pgxpool.Pool) (db.ContestsProblemsRepo, error) {
	ans := &ContestsProblemsRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (c *ContestsProblemsRepoImp) GetContestProblems(ctx context.Context, contestID int64) ([]int64, error) {
	selectContestProblemsStmt := `
	SELECT problem_id FROM contest_problems WHERE contest_id = $1
	`

	rows, err := c.conn.Query(ctx, selectContestProblemsStmt, contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	problemIDs := make([]int64, 0)
	for rows.Next() {
		var problemID int64
		err = rows.Scan(&problemID)
		if err != nil {
			return nil, err
		}

		problemIDs = append(problemIDs, problemID)
	}

	return problemIDs, nil
}

func (c *ContestsProblemsRepoImp) AddProblem(ctx context.Context, contestID int64, problemID int64) error {
	insertContestProblemsStmt := `
		INSERT INTO contest_problems(
			contest_id, problem_id) 
		VALUES($1, $2)
	`
	_, err := c.conn.Exec(ctx, insertContestProblemsStmt, contestID, problemID)
	if err != nil {
		return err
	}
	return nil
}
