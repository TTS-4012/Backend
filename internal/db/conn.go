package db

import (
	"context"
	"github.com/ocontest/backend/internal/db/postgres"
	"github.com/ocontest/backend/internal/db/repos"
	"github.com/ocontest/backend/internal/db/sqlite"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/pkg/errors"
)

type RepoFunc func(context.Context, interface{}) (interface{}, error)
type RepoWrapper func(context.Context, any) error

func pgxWrapper(ctx context.Context, c configs.SectionPostgres) (RepoWrapper, error) {
	pool, err := postgres.NewConnectionPool(ctx, c)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, r any) error {
		if repo, ok := r.(*repos.UsersRepo); ok {
			*repo, err = postgres.NewAuthRepo(ctx, pool)
			return err
		}
		if repo, ok := r.(*repos.ProblemsMetadataRepo); ok {
			*repo, err = postgres.NewProblemsMetadataRepo(ctx, pool)
			return err
		}
		if repo, ok := r.(*repos.SubmissionMetadataRepo); ok {
			*repo, err = postgres.NewSubmissionRepo(ctx, pool)
			return err
		}
		if repo, ok := r.(*repos.TestCaseRepo); ok {
			*repo, err = postgres.NewTestCaseRepo(ctx, pool)
			return err
		}
		if repo, ok := r.(*repos.ContestsMetadataRepo); ok {
			*repo, err = postgres.NewContestsMetadataRepo(ctx, pool)
			return err
		}
		if repo, ok := r.(*repos.ContestsProblemsRepo); ok {
			*repo, err = postgres.NewContestsProblemsMetadataRepo(ctx, pool)
			return err
		}
		if repo, ok := r.(*repos.ContestsUsersRepo); ok {
			*repo, err = postgres.NewContestsUsersRepo(ctx, pool)
			return err
		}
		return errors.WithMessage(pkg.ErrBadRequest, "we don't have an example for your described repo")
	}, nil
}

func sqlWrapper(ctx context.Context, c configs.SectionSQLDB) (RepoWrapper, error) {
	conn, err := sqlite.NewDBConn(ctx, c.DBType, c.ConnUrl)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, r any) error {
		if repo, ok := r.(*repos.UsersRepo); ok {
			*repo, err = sqlite.NewAuthRepo(ctx, conn)
			return err
		}
		if repo, ok := r.(*repos.ProblemsMetadataRepo); ok {
			*repo, err = sqlite.NewProblemsMetadataRepo(ctx, conn)
			return err
		}
		if repo, ok := r.(*repos.SubmissionMetadataRepo); ok {
			*repo, err = sqlite.NewSubmissionRepo(ctx, conn)
			return err
		}
		if repo, ok := r.(*repos.TestCaseRepo); ok {
			*repo, err = sqlite.NewTestCaseRepo(ctx, conn)
			return err
		}
		if repo, ok := r.(*repos.ContestsMetadataRepo); ok {
			*repo, err = sqlite.NewContestsMetadataRepo(ctx, conn)
			return err
		}
		if repo, ok := r.(*repos.ContestsProblemsRepo); ok {
			*repo, err = sqlite.NewContestsProblemsMetadataRepo(ctx, conn)
			return err
		}
		if repo, ok := r.(*repos.ContestsUsersRepo); ok {
			*repo, err = sqlite.NewContestsUsersRepo(ctx, conn)
			return err
		}
		return errors.WithMessage(pkg.ErrBadRequest, "we don't have an example for your described repo")
	}, nil

}
func NewRepoWrapper(ctx context.Context, c configs.SectionSQLDB) (RepoWrapper, error) {
	if c.DBType == "pgx" {
		return pgxWrapper(ctx, c.Postgres)
	}
	return sqlWrapper(ctx, c)
}
