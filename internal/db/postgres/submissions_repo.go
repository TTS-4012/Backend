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

type SubmissionRepoImp struct {
	conn *pgxpool.Pool
}

func NewSubmissionRepo(ctx context.Context, conn *pgxpool.Pool) (db.SubmissionMetadataRepo, error) {
	ans := &SubmissionRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}
func (a *SubmissionRepoImp) Migrate(ctx context.Context) error {
	stmts := []string{
		"CREATE TYPE submission_status AS ENUM('unprocessed', 'processing', 'processed')",
		"CREATE TYPE submission_language AS ENUM('python')",
		`
	CREATE TABLE IF NOT EXISTS submissions(
	    id SERIAL PRIMARY KEY ,
	    problem_id bigint,
	    user_id bigint,
	    file_name varchar(50),
		score int DEFAULT 0,
		status submission_status DEFAULT 'unprocessed',
		language submission_language,
	    created_at TIMESTAMP DEFAULT NOW()
	)`}

	var err error
	for _, s := range stmts {
		_, err = a.conn.Exec(ctx, s)
	}

	return err
}
func (s *SubmissionRepoImp) Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error) {
	stmt := `
	INSERT INTO submissions(problem_id, user_id, file_name, language) VALUES($1, $2, $3, $4)
	`

	var id int64
	err := s.conn.QueryRow(ctx, stmt, submission.ProblemID, submission.UserID, submission.FileName, submission.Language).Scan(&id)
	return id, err
}

func (s *SubmissionRepoImp) Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error) {
	stmt := `
	SELECT id, problem_id, user_id, file_name, score, status, language FROM submissions WHERE id = $1
	`
	var ans structs.SubmissionMetadata
	err := s.conn.QueryRow(ctx, stmt, id).Scan(
		&ans.ID, &ans.ProblemID, &ans.UserID, &ans.FileName, &ans.Score, &ans.Status, &ans.Language)

	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return ans, err
}
