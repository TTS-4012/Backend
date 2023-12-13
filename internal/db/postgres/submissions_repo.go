package postgres

import (
	"context"
	"errors"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
			id SERIAL,
			problem_id bigint not null,
			user_id bigint not null ,
			file_name varchar(50),
			judge_result_id varchar(70),
			status submission_status DEFAULT 'unprocessed',
			language submission_language,
			public boolean DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),

			primary key (id, problem_id, user_id),

			CONSTRAINT fk_problem_id FOREIGN KEY(problem_id) REFERENCES problems(id),
			CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id)
	)`}

	var err error
	for _, s := range stmts {
		_, err = a.conn.Exec(ctx, s)
	}

	return err
}

func (s *SubmissionRepoImp) Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error) {
	stmt := `
	INSERT INTO submissions(
		problem_id, user_id, file_name, language) 
		VALUES($1, $2, $3, $4) RETURNING id
	`

	var id int64
	err := s.conn.QueryRow(ctx, stmt, submission.ProblemID, submission.UserID, submission.FileName, submission.Language).Scan(&id)
	return id, err
}

func (s *SubmissionRepoImp) Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error) {
	stmt := `
	SELECT id, problem_id, user_id, file_name, judge_result_id, status, language, public FROM submissions WHERE id = $1
	`
	var ans structs.SubmissionMetadata
	err := s.conn.QueryRow(ctx, stmt, id).Scan(
		&ans.ID, &ans.ProblemID, &ans.UserID, &ans.FileName, &ans.JudgeResultID, &ans.Status, &ans.Language, &ans.Public)

	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return ans, err
}

func (s *SubmissionRepoImp) AddJudgeResult(ctx context.Context, id int64, docID string) error {
	stmt := `
	UPDATE submissions SET judge_result_id = $1 where id = $2
	`
	_, err := s.conn.Exec(ctx, stmt, docID, id)
	return err
}

func (s *SubmissionRepoImp) ListSubmissions(ctx context.Context, problemID, userID int64, descending bool, limit, offset int, getCount bool) ([]structs.SubmissionMetadata, int, error) {
	panic("implement me!")
}
