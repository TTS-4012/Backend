package postgres

import (
	"context"
	"errors"
	"fmt"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProblemsMetadataRepoImp struct {
	conn *pgxpool.Pool
}

// NOTE: there should probable be an index for searchable columns
var SearchableColumns = map[string]string{
	"solve_count": "solve_count",
	"hardness":    "hardness",
	"problem_id":  "id",
}

func NewProblemsMetadataRepo(ctx context.Context, conn *pgxpool.Pool) (db.ProblemsMetadataRepo, error) {
	ans := &ProblemsMetadataRepoImp{conn: conn}
	return ans, ans.Migrate(ctx)
}

func (a *ProblemsMetadataRepoImp) Migrate(ctx context.Context) (err error) {
	// TODO: remove it if contest module is implemented.
	_, err = a.conn.Exec(ctx, `create table if not exists contests(id bigint primary key);`)
	if err != nil {
		return err
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS problems(
	    id SERIAL PRIMARY KEY ,
	    contest_id bigint,
	    created_by int NOT NULL ,
		title varchar(70) NOT NULL ,
	    document_id varchar(70) NOT NULL ,
	    solve_count int DEFAULT 0,
		hardness int DEFAULT NULL,
	    created_at TIMESTAMP DEFAULT NOW(),
	    CONSTRAINT fk_created_by FOREIGN KEY(created_by) REFERENCES users(id),
		CONSTRAINT fk_contests_id FOREIGN KEY(id) REFERENCES contests(id)
	)
	`

	_, err = a.conn.Exec(ctx, stmt)
	return err
}
func (a *ProblemsMetadataRepoImp) InsertProblem(ctx context.Context, problem structs.Problem) (int64, error) {

	stmt := `
	INSERT INTO problems(
		created_by, title, document_id) 
		VALUES($1, $2, $3) RETURNING id
	`
	var id int64
	err := a.conn.QueryRow(ctx, stmt, problem.CreatedBy, problem.Title, problem.DocumentID).Scan(&id)
	return id, err
}

func (a *ProblemsMetadataRepoImp) GetProblem(ctx context.Context, id int64) (structs.Problem, error) {
	stmt := `
	SELECT created_by, title, document_id, solve_count, coalesce(hardness, -1) FROM problems WHERE id = $1
	`
	var problem structs.Problem
	err := a.conn.QueryRow(ctx, stmt, id).Scan(
		&problem.CreatedBy, &problem.Title, &problem.DocumentID, &problem.SolvedCount, &problem.Hardness)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return problem, err
}

func (a *ProblemsMetadataRepoImp) ListProblems(ctx context.Context, searchColumn string, descending bool, limit, offset int, getCount bool) ([]structs.Problem, int, error) {
	stmt := `
	SELECT id, created_by, title, document_id, solve_count, COALESCE(hardness, -1)
	`
	if getCount {
		stmt = fmt.Sprintf("%s, COUNT(*) OVER() AS total_count", stmt)
	}
	stmt = fmt.Sprintf("%s FROM problems ORDER BY", stmt)

	colName, exists := SearchableColumns[searchColumn]
	if !exists {
		pkg.Log.Warning("tried to list problems with unregistered col name: ", searchColumn)
		return nil, 0, pkg.ErrBadRequest
	}
	stmt += " " + colName
	if descending {
		stmt += " DEC"
	}
	if limit != 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, limit)
	}
	if offset != 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, offset)
	}

	rows, err := a.conn.Query(ctx, stmt)
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

func (a *ProblemsMetadataRepoImp) UpdateProblem(ctx context.Context, id int64, title string) error {
	stmt := `
	UPDATE problems SET title = $1 WHERE id = $2
	`

	_, err := a.conn.Exec(ctx, stmt, title, id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return err
}

func (a *ProblemsMetadataRepoImp) DeleteProblem(ctx context.Context, id int64) (string, error) {
	stmt := `
	DELETE FROM problems WHERE id = $1 RETURNING document_id
	`
	var documentId string
	err := a.conn.QueryRow(ctx, stmt, id).Scan(documentId)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pkg.ErrNotFound
	}
	return documentId, err
}
