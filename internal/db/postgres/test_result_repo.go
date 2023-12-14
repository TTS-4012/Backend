package postgres

import (
	"context"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg/structs"
	"github.com/pkg/errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TestResultRepoImp struct {
	conn *pgxpool.Pool
}

func NewTestResultRepo(ctx context.Context, conn *pgxpool.Pool) (db.TestResultRepo, error) {
	ans := &TestResultRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *TestResultRepoImp) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TYPE test_result_verdict AS ENUM('OK', 'WR', 'TL', 'ML', 'RE', 'XX', 'CE', '??')`,
		`CREATE TABLE IF NOT EXISTS test_results(
			testcase_id bigint not null,
			submission_id bigint not null,
			
			runner_output text,
			runner_error text,

			verdict test_result_verdict,

			primary key (testcase_id, submission_id),

			CONSTRAINT fk_problem_id FOREIGN KEY(problem_id) REFERENCES problems(id),
			CONSTRAINT fk_submission_id FOREIGN KEY(submission_id) REFERENCES submissions(id)
		)`}
	for _, s := range stmts {
		_, err := a.conn.Exec(ctx, s)
		if err != nil {
			return errors.Wrap(err, "error on test result migration")
		}
	}

	return nil
}

func (t *TestResultRepoImp) Insert(ctx context.Context, testResult structs.TestResult) (err error) {
	stmt := `
	INSERT INTO 
		test_results(testcase_id, submission_id, runner_output, runner_error, verdict)
		VALUES($1, $2, $3, $4, %5)`
	_, err = t.conn.Exec(
		ctx, stmt,
		testResult.TestcaseID, testResult.SubmissionID, testResult.RunnerOutput, testResult.RunnerError, testResult.Verdict.String())
	return err
}

// Get not sure if we need it
func (t *TestResultRepoImp) GetByID(ctx context.Context, submissionId int64, testcaseId int64) (ans structs.TestResult, err error) {
	stmt := `
	SELECT testcase_id, submission_id, runner_output, runner_error, verdict FROM test_results
	WHERE submission_id = $1 and testcase_id = $2
	`
	var v string
	err = t.conn.QueryRow(ctx, stmt, submissionId, testcaseId).Scan(&ans.TestcaseID, &ans.SubmissionID, &ans.RunnerOutput, &ans.RunnerError, &v)
	ans.Verdict = structs.VerdictFromString(v)
	return ans, err
}
