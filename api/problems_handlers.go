package api

import (
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *handlers) CreateProblem(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "createProblem")

	var reqData structs.RequestCreateProblem
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.problemsHandler.CreateProblem(c, reqData)
	if status == http.StatusOK && reqData.ContestID != 0 {
		var newReq structs.RequestAddProblemContest
		newReq.ProblemID = resp.ProblemID
		newReq.ContestID = reqData.ContestID
		status = h.contestsHandler.AddProblemContest(c, newReq)
	}
	c.JSON(status, resp)
}

func (h *handlers) GetProblem(c *gin.Context) {
	//logger := pkg.Log.WithField("handler", "getProblem")

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}
	resp, status := h.problemsHandler.GetProblem(c, problemID)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) ListProblems(c *gin.Context) {

	logger := pkg.Log.WithField("handler", "listProblem")
	var reqData structs.RequestListProblems

	reqData.OrderedBy = c.Query("ordered_by")
	if reqData.OrderedBy == "" {
		reqData.OrderedBy = "problem_id"
	}
	reqData.Descending = c.Query("descending") == "true"

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

	resp, status := h.problemsHandler.ListProblem(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) UpdateProblem(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "updateProblem")

	var reqData structs.RequestUpdateProblem
	var err error
	reqData.Id, err = strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("Failed to parse id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	status := h.problemsHandler.UpdateProblem(c, reqData)
	c.Status(status)
}

func (h *handlers) DeleteProblem(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "deleteProblem")

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("Failed to parse id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	status := h.problemsHandler.DeleteProblem(c, problemID)
	c.Status(status)
}
