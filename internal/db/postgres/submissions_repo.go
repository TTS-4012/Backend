package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/ocontest/backend/internal/db/repos"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepoImp struct {
	conn *pgxpool.Pool
}

func NewSubmissionRepo(ctx context.Context, conn *pgxpool.Pool) (repos.SubmissionMetadataRepo, error) {
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
			user_id bigint not null,
			contest_id bigint,
			file_name varchar(50),
			judge_result_id varchar(70) DEFAULT '',
			score int DEFAULT 0,
			status submission_status DEFAULT 'unprocessed',
			language submission_language,
			is_final boolean DEFAULT FALSE,
			public boolean DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),

			unique(id),
			primary key (id, problem_id, user_id),

			CONSTRAINT fk_problem_id FOREIGN KEY(problem_id) REFERENCES problems(id),
			CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id),
			CONSTRAINT fk_contest_id FOREIGN KEY(contest_id) REFERENCES contests(id)
	)`}

	var err error
	for _, s := range stmts {
		_, err = a.conn.Exec(ctx, s)
	}

	return err
}

func (s *SubmissionRepoImp) Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error) {
	stmt := `
	INSERT INTO submissions(problem_id, user_id, file_name, language
	`
	if submission.ContestID != 0 {
		stmt += ", contest_id) VALUES ($1, $2, $3, $4, $5)"
	} else {
		stmt += ") VALUES ($1, $2, $3, $4)"
	}
	stmt += " RETURNING id"

	var id int64
	var err error
	if submission.ContestID != 0 {
		err = s.conn.QueryRow(ctx, stmt, submission.ProblemID, submission.UserID, submission.FileName, submission.Language, submission.ContestID).Scan(&id)
	} else {
		err = s.conn.QueryRow(ctx, stmt, submission.ProblemID, submission.UserID, submission.FileName, submission.Language).Scan(&id)
	}

	pkg.Log.Debug(err)
	return id, err
}

func (s *SubmissionRepoImp) Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error) {
	stmt := `
	SELECT id, problem_id, user_id, coalesce(contest_id, 0), file_name, score, coalesce(judge_result_id, ''), status, language, is_final, public, created_at FROM submissions WHERE id = $1
	`
	var ans structs.SubmissionMetadata
	var t time.Time
	err := s.conn.QueryRow(ctx, stmt, id).Scan(
		&ans.ID, &ans.ProblemID, &ans.UserID, &ans.ContestID, &ans.FileName, &ans.Score, &ans.JudgeResultID, &ans.Status, &ans.Language, &ans.IsFinal, &ans.Public, &t)

	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	ans.CreatedAT = t.Format(time.RFC3339)
	return ans, err
}

func (s *SubmissionRepoImp) GetByProblem(ctx context.Context, problemID int64) ([]structs.SubmissionMetadata, error) {
	stmt := `
	SELECT 
		id, problem_id, user_id, coalesce(contest_id, 0), file_name, score, coalesce(judge_result_id, ''),
			status, language, is_final, public, created_at 
		FROM submissions WHERE problem_id = $1 and is_final = true
	`

	rows, err := s.conn.Query(ctx, stmt, problemID)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkg.ErrNotFound
		}
		return nil, errors.WithStack(err)

	}

	var t time.Time
	ans := make([]structs.SubmissionMetadata, 0)
	for rows.Next() {
		var row structs.SubmissionMetadata
		err = rows.Scan(&row.ID, &row.ProblemID, &row.UserID, &row.ContestID, &row.FileName, &row.Score, &row.JudgeResultID, &row.Status, &row.Language, &row.IsFinal, &row.Public, &t)
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
		id, problem_id, user_id, coalesce(contest_id, 0), file_name, score, coalesce(judge_result_id, ''),
			status, language, is_final, public, created_at 
		FROM submissions WHERE is_final = true AND problem_id = $1 AND user_id = $2
	`
	if contestID != 0 {
		stmt += " AND contest_id = $3"
	}

	var ans structs.SubmissionMetadata
	var t time.Time
	var err error
	if contestID != 0 {
		err = s.conn.QueryRow(ctx, stmt, problemID, userID, contestID).Scan(&ans.ID, &ans.ProblemID, &ans.UserID, &ans.ContestID, &ans.FileName, &ans.Score, &ans.JudgeResultID, &ans.Status, &ans.Language, &ans.IsFinal, &ans.Public, &t)
	} else {
		err = s.conn.QueryRow(ctx, stmt, problemID, userID).Scan(&ans.ID, &ans.ProblemID, &ans.UserID, &ans.ContestID, &ans.FileName, &ans.Score, &ans.JudgeResultID, &ans.Status, &ans.Language, &ans.IsFinal, &ans.Public, &t)
	}
	ans.CreatedAT = t.Format(time.RFC3339)
	if errors.Is(err, pgx.ErrNoRows) {
		return ans, pkg.ErrNotFound
	}
	return ans, err
}

// UpdateJudgeResults will add judge_result_id, update status, and change is final
func (s *SubmissionRepoImp) UpdateJudgeResults(ctx context.Context, problemID, userID, contestID, submissionID int64, docID string, score int, isFinal bool) error {

	stmt := `
	UPDATE submissions SET is_final = false WHERE problem_id = $1 AND user_id = $2
	`
	if contestID != 0 {
		stmt += " AND contest_id = $3"
	}

	var err error
	if isFinal {
		if contestID != 0 {
			_, err = s.conn.Exec(ctx, stmt, problemID, userID, contestID)
		} else {
			_, err = s.conn.Exec(ctx, stmt, problemID, userID)
		}
		if err != nil {
			return err
		}
	}

	stmt = `
	UPDATE submissions SET status = 'processed', score = $1, judge_result_id = $2, is_final = $3 WHERE id = $4
	`
	_, err = s.conn.Exec(ctx, stmt, score, docID, isFinal, submissionID)
	return err
}

func (s *SubmissionRepoImp) ListSubmissions(ctx context.Context, problemID, userID, contestID int64, descending bool, limit, offset int, getCount bool) ([]structs.SubmissionMetadata, int, error) {
	args := make([]interface{}, 0)

	stmt := `
	SELECT submissions.id, submissions.problem_id, submissions.user_id, coalesce(submissions.contest_id, 0),
	 submissions.file_name, submissions.score, coalesce(submissions.judge_result_id, ''),
	 submissions.status, submissions.language, submissions.public, submissions.created_at, problems.title
	`
	if getCount {
		stmt = fmt.Sprintf("%s, COUNT(*) OVER() AS total_count", stmt)
	}
	stmt += " FROM submissions JOIN problems ON submissions.problem_id = problems.id"

	var cond string
	if problemID != 0 {
		args = append(args, problemID)
		cond = fmt.Sprintf("%s submissions.problem_id = $%d", cond, len(args))
	}
	if userID != 0 {
		if len(args) != 0 {
			cond += " AND"
		}
		args = append(args, userID)
		cond = fmt.Sprintf("%s submissions.user_id = $%d", cond, len(args))
	}
	if contestID != 0 {
		if len(args) != 0 {
			cond += " AND"
		}
		args = append(args, contestID)
		cond = fmt.Sprintf("%s submissions.contest_id = $%d", cond, len(args))
	} else {
		if len(args) != 0 {
			cond += " AND"
		}
		cond = fmt.Sprintf("%s submissions.contest_id IS NULL", cond)
	}
	if len(args) != 0 {
		stmt = fmt.Sprintf("%s WHERE%s", stmt, cond)
	}

	stmt = fmt.Sprintf("%s ORDER BY submissions.created_at", stmt)
	if descending {
		stmt += " DESC"
	}
	if limit != 0 {
		args = append(args, limit)
		stmt = fmt.Sprintf("%s LIMIT $%d", stmt, len(args))
	}
	if offset != 0 {
		args = append(args, offset)
		stmt = fmt.Sprintf("%s OFFSET $%d", stmt, len(args))
	}

	rows, err := s.conn.Query(ctx, stmt, args...)

	if err != nil {
		return nil, 0, err
	}

	ans := make([]structs.SubmissionMetadata, 0)
	var total_count int = 0
	for rows.Next() {
		var submission structs.SubmissionMetadata
		var t time.Time
		if getCount {
			err = rows.Scan(&submission.ID, &submission.ProblemID, &submission.UserID, &submission.ContestID,
				&submission.FileName, &submission.Score, &submission.JudgeResultID, &submission.Status,
				&submission.Language, &submission.Public, &t, &submission.ProblemTitle, &total_count)
		} else {
			err = rows.Scan(&submission.ID, &submission.ProblemID, &submission.UserID, &submission.ContestID,
				&submission.FileName, &submission.Score, &submission.JudgeResultID, &submission.Status,
				&submission.Language, &submission.Public, &t, &submission.ProblemTitle)
		}
		if err != nil {
			return nil, 0, err
		}
		submission.CreatedAT = t.Format(time.RFC3339)
		ans = append(ans, submission)
	}
	return ans, total_count, err
}
