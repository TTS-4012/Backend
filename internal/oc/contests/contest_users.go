package contests

import (
	"context"
	"errors"
	"github.com/ocontest/backend/pkg"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (c ContestsHandlerImp) RegisterUser(ctx context.Context, contestID, userID int64) int {
	logger := pkg.Log.WithFields(logrus.Fields{
		"module": "contest",
		"method": "RegisterUser",
	})

	err := c.contestsUsersRepo.Add(ctx, contestID, userID)
	if err != nil {
		logger.Error("error on insert to db: ", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (c ContestsHandlerImp) UnregisterUser(ctx context.Context, contestID, userID int64) int {
	logger := pkg.Log.WithFields(logrus.Fields{
		"module": "contest",
		"method": "UnregisterUser",
	})

	err := c.contestsUsersRepo.Delete(ctx, contestID, userID)
	if err != nil {
		logger.Error("error on insert to db: ", err)
		if errors.Is(err, pkg.ErrNotFound) {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
