package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"ocontest/internal/db"
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
	return 0, nil
}

func (c *ContestsMetadataRepoImp) GetContest(ctx context.Context, id int64) (structs.Contest, error) {
	return structs.Contest{}, nil
}
func (c *ContestsMetadataRepoImp) ListContests(ctx context.Context) ([]structs.Contest, error) {
	return nil, nil
}
