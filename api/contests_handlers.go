package api

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
)

func (h *handlers) CreateContest(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "createContest")

	var reqData structs.RequestCreateContest
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.contestsHandler.CreateContest(c, reqData)
	c.JSON(status, resp)
}

func (h *handlers) GetContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}
	resp, status := h.contestsHandler.GetContest(c, contestID)
	if status != http.StatusOK {
		c.Status(status)
		return
	}
	c.JSON(status, resp)
}

func (h *handlers) ListContests(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "listContests")

	var reqData structs.RequestListContests

	reqData.Descending = c.Query("descending") == "true"
	reqData.Started = c.Query("started") == "true"

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var errLimit, errOffset error
	if limitStr != "" {
		reqData.Limit, errLimit = strconv.Atoi(limitStr)
	}
	if offsetStr != "" {
		reqData.Offset, errOffset = strconv.Atoi(offsetStr)
	}
	if errLimit != nil || errOffset != nil {
		logger.Warningf("invalid limit and/or offset, limit: %v offset: %v", limitStr, offsetStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid limit or offset, limit and offset should be integers",
		})
		return
	}

	reqData.MyContest = c.Query("my_contest") == "true"

	resp, status := h.contestsHandler.ListContests(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) UpdateContest(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "updateContest")

	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("Failed to parse id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	var reqData structs.RequestUpdateContest
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	status := h.contestsHandler.UpdateContest(c, contestID, reqData)
	c.Status(status)
}

func (h *handlers) DeleteContest(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "deleteContest")

	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		logger.Warn("Failed to parse id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	status := h.contestsHandler.DeleteContest(c, contestID)
	c.Status(status)
}

func (h *handlers) AddProblemContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid contest id, id should be an integer",
		})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem id, id should be an integer",
		})
		return
	}

	status := h.contestsHandler.AddProblemToContest(c, contestID, problemID)
	c.Status(status)
}

func (h *handlers) RemoveProblemContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid contest id, id should be an integer",
		})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem id, id should be an integer",
		})
		return
	}

	status := h.contestsHandler.RemoveProblemFromContest(c, contestID, problemID)
	c.Status(status)
}

func (h *handlers) GetContestScoreboard(c *gin.Context) {
	logger := logrus.WithField("handler", "GetContestScoreboard")
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	var reqData structs.RequestGetScoreboard
	reqData.ContestID = contestID

	reqData.GetCount = c.Query("get_count") == "true"

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	var errLimit, errOffset error
	if limitStr != "" {
		reqData.Limit, errLimit = strconv.Atoi(limitStr)
	}
	if offsetStr != "" {
		reqData.Offset, errOffset = strconv.Atoi(offsetStr)
	}
	if errLimit != nil || errOffset != nil {
		logger.Warningf("invalid limit and/or offset, limit: %v offset: %v", limitStr, offsetStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid limit or offset, limit and offset should be integers",
		})
		return
	}

	resp, status := h.contestsHandler.GetContestScoreboard(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) PatchContest(c *gin.Context) {
	action := c.Query("action")

	userID := c.Value("user_id").(int64)
	contestID := c.Value("user_id").(int64)
	pkg.Log.Debug(userID, contestID)
	switch action {
	case "register":
		c.Status(h.contestsHandler.RegisterUser(c, contestID, userID))
	case "unregister":
		c.Status(h.contestsHandler.UnregisterUser(c, contestID, userID))
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action " + action + " not defined",
		})
	}
}
