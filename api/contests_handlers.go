package api

import (
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
	_ = pkg.Log.WithField("handler", "getContest")

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
	problemIDs, status := h.contestsHandler.GetContestProblems(c, contestID)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	for _, problemID := range problemIDs {
		problem, status := h.problemsHandler.GetProblem(c, problemID)
		if status != http.StatusOK {
			c.Status(status)
			return
		}

		resp.Problems = append(resp.Problems, structs.ContestProblem{
			ID:    problem.ProblemID,
			Title: problem.Title,
		})
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

	resp, status := h.contestsHandler.ListContests(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) UpdateContest(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *handlers) DeleteContest(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *handlers) AddProblemContest(c *gin.Context) {
	_ = pkg.Log.WithField("handler", "addProblemContest")

	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid contest id, id should be an integer",
		})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	_ = pkg.Log.WithField("handler", "removeProblemContest")

	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid contest id, id should be an integer",
		})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	resp, status := h.contestsHandler.GetContestScoreboard(c, contestID)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}
