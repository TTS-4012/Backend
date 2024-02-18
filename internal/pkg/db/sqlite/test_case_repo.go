package sqlite

import (
	"context"
	"database/sql"

	"github.com/ocontest/backend/internal/pkg/db/repos"

	"github.com/ocontest/backend/pkg/structs"
	"github.com/pkg/errors"
)

type TestCaseRepoImp struct {
	conn *sql.DB
}

func NewTestCaseRepo(ctx context.Context, conn *sql.DB) (repos.TestCaseRepo, error) {
	ans := &TestCaseRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *TestCaseRepoImp) Migrate(ctx context.Context) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS testcases(
			id INTEGER AUTO_INCREMENT,
			problem_id bigint not null,
			
			input text not null ,
			output text not null ,

			unique(id),
			primary key (problem_id, id),

			CONSTRAINT fk_problem_id FOREIGN KEY(problem_id) REFERENCES problems(id)
	)`

	_, err := a.conn.ExecContext(ctx, stmt)

	return err
}

func (t *TestCaseRepoImp) Insert(ctx context.Context, testCase structs.Testcase) (id int64, err error) {
	stmt := `INSERT INTO testcases(problem_id, input, output) VALUES($, $, $) RETURNING id`
	err = t.conn.QueryRowContext(ctx, stmt, testCase.ProblemID, testCase.Input, testCase.ExpectedOutput).Scan(&id)
	if err != nil {
		err = errors.Wrap(err, "error on inserting to testcase repos")
		return
	}
	return id, nil
}

// Get not sure if we need it
func (t *TestCaseRepoImp) GetByID(ctx context.Context, id int64) (ans structs.Testcase, err error) {
	stmt := `
	SELECT id, problem_id, input, output FROM testcases WHERE id = $
	`
	err = t.conn.QueryRowContext(ctx, stmt, id).Scan(&ans.ID, &ans.ProblemID, &ans.Input, &ans.ExpectedOutput)
	return ans, err
}

// GetAllTestsOfProblem since our first part of primary key is problem id, there will be no performance issue
func (t *TestCaseRepoImp) GetAllTestsOfProblem(ctx context.Context, problemID int64) ([]structs.Testcase, error) {
	stmt := `
	SELECT id, problem_id, input, output FROM testcases WHERE problem_id = $
	`
	rows, err := t.conn.QueryContext(ctx, stmt, problemID)
	if err != nil {
		err = errors.Wrap(err, "error on executing query on pg")
		err = errors.WithStack(err)
		return nil, err
	}
	defer rows.Close()
	ans := make([]structs.Testcase, 0)
	for rows.Next() {
		var newTestcase structs.Testcase
		err := rows.Scan(&newTestcase.ID, &newTestcase.ProblemID, &newTestcase.Input, &newTestcase.ExpectedOutput)
		if err != nil {
			err = errors.Wrap(err, "error on reading row")
			return nil, errors.WithStack(err)
		}
		ans = append(ans, newTestcase)
	}

	return ans, nil
}
