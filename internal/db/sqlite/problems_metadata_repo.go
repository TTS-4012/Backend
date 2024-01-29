package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/ocontest/backend/internal/db/repos"

	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
)

type ProblemsMetadataRepoImp struct {
	conn *sql.DB
}

var SearchableColumns = map[string]string{
	"solve_count": "solve_count",
	"hardness":    "hardness",
	"problem_id":  "id",
}

func NewProblemsMetadataRepo(ctx context.Context, conn *sql.DB) (repos.ProblemsMetadataRepo, error) {
	ans := &ProblemsMetadataRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *ProblemsMetadataRepoImp) Migrate(ctx context.Context) (err error) {
	stmt := []string{`
	CREATE TABLE IF NOT EXISTS problems(
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    created_by int NOT NULL ,
		title varchar(70) NOT NULL ,
	    document_id varchar(70) NOT NULL ,
	    solve_count int DEFAULT 0,
		hardness int DEFAULT NULL,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	is_private BOOL NOT NULL DEFAULT FALSE,
	   FOREIGN KEY(created_by) REFERENCES users(id)
	)
	`}
	for _, s := range stmt {
		_, err = a.conn.ExecContext(ctx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ProblemsMetadataRepoImp) InsertProblem(ctx context.Context, problem structs.Problem) (int64, error) {

	stmt := `
	INSERT INTO problems(
		created_by, title, document_id, hardness, is_private) 
		VALUES($, $, $, $, $) RETURNING id
	`
	var id int64
	err := a.conn.QueryRowContext(ctx, stmt, problem.CreatedBy, problem.Title, problem.DocumentID, problem.Hardness, problem.IsPrivate).Scan(&id)
	return id, err
}

func (a *ProblemsMetadataRepoImp) GetProblem(ctx context.Context, id int64) (structs.Problem, error) {
	stmt := `
	SELECT created_by, title, document_id, solve_count, coalesce(hardness, -1) FROM problems WHERE id = $
	`
	var problem structs.Problem
	err := a.conn.QueryRowContext(ctx, stmt, id).Scan(
		&problem.CreatedBy, &problem.Title, &problem.DocumentID, &problem.SolvedCount, &problem.Hardness)
	if errors.Is(err, sql.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return problem, err
}

func (a *ProblemsMetadataRepoImp) GetProblemTitle(ctx context.Context, id int64) (string, error) {
	stmt := `
	SELECT title FROM problems WHERE id = $
	`
	var ans string
	err := a.conn.QueryRowContext(ctx, stmt, id).Scan(&ans)
	if errors.Is(err, sql.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return ans, err
}

func (a *ProblemsMetadataRepoImp) ListProblems(ctx context.Context, searchColumn string, descending bool, limit, offset int, getCount bool) ([]structs.Problem, int, error) {
	stmt := `
	SELECT id, created_by, title, document_id, solve_count, COALESCE(hardness, -1)
	`
	if getCount {
		stmt = fmt.Sprintf("%s, COUNT(*) OVER() AS total_count", stmt)
	}
	stmt = fmt.Sprintf("%s FROM problems WHERE is_private = false ORDER BY", stmt)

	colName, exists := SearchableColumns[searchColumn]
	if !exists {
		pkg.Log.Warning("tried to list problems with unregistered col name: ", searchColumn)
		return nil, 0, pkg.ErrBadRequest
	}
	stmt += " " + colName
	if descending {
		stmt += " DESC"
	}
	if limit != 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, limit)
	}
	if offset != 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, offset)
	}

	rows, err := a.conn.QueryContext(ctx, stmt)
	if err != nil {
		return nil, 0, err
	}

	ans := make([]structs.Problem, 0)
	var total_count int = 0
	for rows.Next() {

		var problem structs.Problem
		if getCount {
			err = rows.Scan(&problem.ID, &problem.CreatedBy, &problem.Title, &problem.DocumentID, &problem.SolvedCount, &problem.Hardness, &total_count)
		} else {
			err = rows.Scan(&problem.ID, &problem.CreatedBy, &problem.Title, &problem.DocumentID, &problem.SolvedCount, &problem.Hardness)
		}
		if err != nil {
			return nil, 0, err
		}
		ans = append(ans, problem)
	}
	return ans, total_count, err
}

// TODO: suitable query builder yada yada
func (a *ProblemsMetadataRepoImp) UpdateProblem(ctx context.Context, id int64, title string, hardness int64) error {
	stmt := `
	UPDATE problems SET
	`

	args := make([]interface{}, 0)

	if title != "" {
		args = append(args, title)
		stmt += " title = $"
	}
	if hardness != 0 {
		if len(args) > 0 {
			stmt += ","
		}
		args = append(args, hardness)
		stmt += " hardness = $"
	}
	if len(args) == 0 {
		return nil
	}

	args = append(args, id)
	stmt = fmt.Sprintf("%s WHERE id = $%d", stmt, len(args))

	_, err := a.conn.ExecContext(ctx, stmt, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}

func (a *ProblemsMetadataRepoImp) DeleteProblem(ctx context.Context, id int64) (string, error) {
	stmt := `
	DELETE FROM problems WHERE id = $ RETURNING document_id
	`
	var documentId string
	err := a.conn.QueryRowContext(ctx, stmt, id).Scan(documentId)
	if errors.Is(err, sql.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return documentId, err
}
