package api

import (
	"net/http"
	"strconv"

	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"

	"github.com/gin-gonic/gin"
)

func (h *handlers) Submit(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "Submit")

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError.Error(),
		})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting problem_id from url: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem_id, problem_id should be an integer",
		})
		return
	}

	file_name := c.GetHeader("Filename")
	if file_name == "" {
		file_name = "filename"
	}

	buffer, err := c.GetRawData()
	if err != nil {
		logger.Error("Failed reading file from request body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError.Error(),
		})
		return
	}

	var contestID int64
	contestIDStr := c.Query("contest_id")
	if contestIDStr == "" {
		contestID = 0
	} else {
		contestID, err = strconv.ParseInt(contestIDStr, 10, 64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "coudn't get contestID",
			})
		}
	}

	reqData := structs.RequestSubmit{
		UserID:      userID.(int64),
		ProblemID:   problemID,
		ContestID:   contestID,
		Code:        buffer,
		FileName:    file_name,
		ContentType: c.GetHeader("Content-Type"),
		Language:    "python", // For now just python
	}

	submissionID, status := h.submissionsHandler.Submit(c, reqData)
	c.JSON(status, gin.H{
		"submission_id": submissionID,
	})
}

func (h *handlers) GetSubmission(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "GetSubmission")

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError.Error(),
		})
		return
	}

	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting id from request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	setDownloadHeaders := c.Query("download") == "true"

	resp, _, status := h.submissionsHandler.Get(c, userID.(int64), submissionID)
	if status == http.StatusOK {
		if setDownloadHeaders {
			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", "attachment; filename="+resp.Metadata.FileName)
		}
		c.Data(status, "application/octet-stream", resp.RawCode)
	} else {
		c.Status(status)
	}
}

func (h *handlers) GetSubmissionResult(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "GetSubmissionResult")

	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting id from request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	resp, status := h.submissionsHandler.GetResults(c, submissionID)

	c.JSON(status, resp)
}

func (h *handlers) ListSubmissions(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "ListSubmissions")
	var reqData structs.RequestListSubmissions

	reqData.UserID = 0
	if c.Query("get_all") != "true" {
		userID, exists := c.Get(UserIDKey)
		if !exists {
			logger.Error("error on getting user_id from context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": pkg.ErrInternalServerError.Error(),
			})
			return
		}
		reqData.UserID = userID.(int64)
	}

	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting problem_id from url: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem_id, problem_id should be an integer",
		})
		return
	}
	reqData.ProblemID = problemID

	reqData.ContestID = 0

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

	resp, status := h.submissionsHandler.ListSubmission(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) ListContestSubmissions(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "ListContestSubmissions")
	var reqData structs.RequestListSubmissions

	reqData.UserID = 0
	if c.Query("get_all") != "true" {
		userID, exists := c.Get(UserIDKey)
		if !exists {
			logger.Error("error on getting user_id from context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": pkg.ErrInternalServerError.Error(),
			})
			return
		}
		reqData.UserID = userID.(int64)
	}

	reqData.ProblemID = 0

	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting contest_id from url: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid contest_id, contest_id should be an integer",
		})
		return
	}
	reqData.ContestID = contestID

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

	resp, status := h.submissionsHandler.ListSubmission(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}

func (h *handlers) ListContestProblemSubmissions(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "ListContestProblemSubmissions")
	var reqData structs.RequestListSubmissions

	reqData.UserID = 0
	if c.Query("get_all") != "true" {
		userID, exists := c.Get(UserIDKey)
		if !exists {
			logger.Error("error on getting user_id from context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": pkg.ErrInternalServerError.Error(),
			})
			return
		}
		reqData.UserID = userID.(int64)
	}

	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil {
		logger.Error("error on getting problem_id from url: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem_id, problem_id should be an integer",
		})
		return
	}
	reqData.ProblemID = problemID

	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting contest_id from url: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid contest_id, contest_id should be an integer",
		})
		return
	}
	reqData.ContestID = contestID

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

	resp, status := h.submissionsHandler.ListSubmission(c, reqData)
	if status == http.StatusOK {
		c.JSON(status, resp)
	} else {
		c.Status(status)
	}
}
