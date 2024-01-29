package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ocontest/backend/internal/db/repos"
	"time"

	"github.com/pkg/errors"

	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
)

type SubmissionRepoImp struct {
	conn *sql.DB
}

func NewSubmissionRepo(ctx context.Context, conn *sql.DB) (repos.SubmissionMetadataRepo, error) {
	ans := &SubmissionRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *SubmissionRepoImp) Migrate(ctx context.Context) error {
	stmts := []string{
		"CREATE TYPE submission_status AS ENUM('unprocessed', 'processing', 'processed')",
		"CREATE TYPE submission_language AS ENUM('python')",
		`
		CREATE TABLE IF NOT EXISTS submissions(
			id INTEGER AUTO_INCREMENT,
			problem_id bigint not null,
			user_id bigint not null ,
			file_name varchar(50),
			judge_result_id varchar(70) default '',
			score int DEFAULT 0,
			status submission_status DEFAULT 'unprocessed',
			language submission_language,
			is_final boolean DEFAULT FALSE,
			public boolean DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			unique(id),
			primary key (id, problem_id, user_id),

			CONSTRAINT fk_problem_id FOREIGN KEY(problem_id) REFERENCES problems(id),
			CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id)
	)`}

	var err error
	for _, s := range stmts {
		_, err = a.conn.ExecContext(ctx, s)
	}

	return err
}

func (s *SubmissionRepoImp) Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error) {
	stmt := `
	INSERT INTO submissions(
		problem_id, user_id, file_name, language) 
		VALUES($, $, $, $) RETURNING id
	`

	var id int64
	err := s.conn.QueryRowContext(ctx, stmt, submission.ProblemID, submission.UserID, submission.FileName, submission.Language).Scan(&id)
	pkg.Log.Debug(err)
	return id, err
}

func (s *SubmissionRepoImp) Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error) {
	stmt := `
	SELECT id, problem_id, user_id, file_name, score, coalesce(judge_result_id, ''), status, language, is_final, public, created_at FROM submissions WHERE id = $
	`
	var ans structs.SubmissionMetadata
	var t time.Time
	err := s.conn.QueryRowContext(ctx, stmt, id).Scan(
		&ans.ID, &ans.ProblemID, &ans.UserID, &ans.FileName, &ans.Score, &ans.JudgeResultID, &ans.Status, &ans.Language, &ans.IsFinal, &ans.Public, &t)

	if errors.Is(err, sql.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	ans.CreatedAT = t.Format(time.RFC3339)
	return ans, err
}

func (s *SubmissionRepoImp) GetByProblem(ctx context.Context, problemID int64) ([]structs.SubmissionMetadata, error) {
	stmt := `
	SELECT 
		id, problem_id, user_id, file_name, score, coalesce(judge_result_id, ''),
			status, language, is_final, public, created_at 
		FROM submissions WHERE problem_id = $ and is_final = true
	`

	rows, err := s.conn.QueryContext(ctx, stmt, problemID)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkg.ErrNotFound
		}
		return nil, errors.WithStack(err)

	}

	var t time.Time
	ans := make([]structs.SubmissionMetadata, 0)
	for rows.Next() {
		var row structs.SubmissionMetadata
		err = rows.Scan(&row.ID, &row.ProblemID, &row.UserID, &row.FileName, &row.Score, &row.JudgeResultID, &row.Status, &row.Language, &row.IsFinal, &row.Public, &t)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		row.CreatedAT = t.Format(time.RFC3339)

		ans = append(ans, row)
	}

	return ans, nil
}

func (s *SubmissionRepoImp) GetFinalSubmission(ctx context.Context, problemID, userID, contestID int64) (structs.SubmissionMetadata, error) {
	stmt := `
	SELECT 
		id, problem_id, user_id, file_name, score, coalesce(judge_result_id, ''),
			status, language, is_final, public, created_at 
		FROM submissions WHERE user_id = $ and problem_id = $ and contest_id = $ and is_final = true
	`

	var ans structs.SubmissionMetadata
	var t time.Time
	err := s.conn.QueryRowContext(ctx, stmt, userID, problemID, contestID).Scan(&ans.ID, &ans.ProblemID, &ans.UserID, &ans.FileName, &ans.Score, &ans.JudgeResultID, &ans.Status, &ans.Language, &ans.IsFinal, &ans.Public, &t)
	ans.CreatedAT = t.Format(time.RFC3339)
	if errors.Is(err, sql.ErrNoRows) {
		return ans, pkg.ErrNotFound
	}
	return ans, err
}

// UpdateJudgeResults will add judge_result_id, update status, and change is final
// UpdateJudgeResults will add judge_result_id, update status, and change is final
func (s *SubmissionRepoImp) UpdateJudgeResults(ctx context.Context, problemID, userID, contestID, submissionID int64, docID string, score int, isFinal bool) error {

	stmt := `
	UPDATE submissions SET is_final = false WHERE problem_id = $ AND user_id = $
	`
	if contestID != 0 {
		stmt += " AND contest_id = $3"
	}

	var err error
	if isFinal {
		if contestID != 0 {
			_, err = s.conn.ExecContext(ctx, stmt, problemID, userID, contestID)
		} else {
			_, err = s.conn.ExecContext(ctx, stmt, problemID, userID)
		}
		if err != nil {
			return err
		}
	}

	stmt = `
	UPDATE submissions SET status = 'processed', score = $, judge_result_id = $, is_final = $ WHERE id = $
	`
	_, err = s.conn.ExecContext(ctx, stmt, score, docID, isFinal, submissionID)
	return err
}

func (s *SubmissionRepoImp) ListSubmissions(ctx context.Context, problemID, userID, contestID int64, descending bool, limit, offset int, getCount bool) ([]structs.SubmissionMetadata, int, error) {
	args := make([]interface{}, 0)

	stmt := `
	SELECT id, problem_id, user_id, coalesce(contest_id, 0), file_name, score, coalesce(judge_result_id, ''), status, language, public, created_at
	`
	if getCount {
		stmt = fmt.Sprintf("%s, COUNT(*) OVER() AS total_count", stmt)
	}
	stmt = fmt.Sprintf("%s FROM submissions", stmt)

	var cond string
	if problemID != 0 {
		args = append(args, problemID)
		cond = fmt.Sprintf("%s problem_id = $%d", cond, len(args))
	}
	if userID != 0 {
		if len(args) != 0 {
			cond += " AND"
		}
		args = append(args, userID)
		cond = fmt.Sprintf("%s user_id = $%d", cond, len(args))
	}
	if contestID != 0 {
		if len(args) != 0 {
			cond += " AND"
		}
		args = append(args, contestID)
		cond = fmt.Sprintf("%s contest_id = $%d", cond, len(args))
	} else {
		if len(args) != 0 {
			cond += " AND"
		}
		cond = fmt.Sprintf("%s contest_id IS NULL", cond)
	}
	if len(args) != 0 {
		stmt = fmt.Sprintf("%s WHERE %s", stmt, cond)
	}

	stmt = fmt.Sprintf("%s ORDER BY created_at", stmt)
	if descending {
		stmt += " DESC"
	}
	if limit != 0 {
		args = append(args, limit)
		stmt += " LIMIT $"
	}
	if offset != 0 {
		args = append(args, offset)
		stmt += " OFFSET $"
	}

	rows, err := s.conn.QueryContext(ctx, stmt, args...)

	if err != nil {
		return nil, 0, err
	}

	ans := make([]structs.SubmissionMetadata, 0)
	var total_count int = 0
	for rows.Next() {
		var submission structs.SubmissionMetadata
		var t time.Time
		if getCount {
			err = rows.Scan(&submission.ID, &submission.ProblemID, &submission.UserID, &submission.ContestID, &submission.FileName, &submission.Score, &submission.JudgeResultID, &submission.Status, &submission.Language, &submission.Public, &t, &total_count)
		} else {
			err = rows.Scan(&submission.ID, &submission.ProblemID, &submission.UserID, &submission.ContestID, &submission.FileName, &submission.Score, &submission.JudgeResultID, &submission.Status, &submission.Language, &submission.Public, &t)
		}
		if err != nil {
			return nil, 0, err
		}
		submission.CreatedAT = t.Format(time.RFC3339)
		ans = append(ans, submission)
	}
	return ans, total_count, err
}
