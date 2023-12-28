package api

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"

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

	reqData.IsPrivate = false
	if reqData.ContestID != 0 {
		reqData.IsPrivate = true
	}

	resp, status := h.problemsHandler.CreateProblem(c, reqData)
	if status == http.StatusOK && reqData.ContestID != 0 {
		status = h.contestsHandler.AddProblemToContest(c, reqData.ContestID, resp.ProblemID)
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

func (h *handlers) AddTestCase(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "addTestcase")

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("Failed to parse id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	logger.Debug(c.GetHeader("Content-Length"))
	body := c.Request.Body
	buff := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buff, body)
	if err != nil {
		logger.Error("error on create io ReaderAt")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "something went wrong",
		})
		return
	}

	status := h.problemsHandler.AddTestcase(c, problemID, buff.Bytes())
	c.Status(status)
}
