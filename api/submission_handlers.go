package api

import (
	"fmt"
	"net/http"
	"ocontest/pkg"
	"ocontest/pkg/structs"
	"strconv"

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

	buffer, err := c.GetRawData()
	if err != nil {
		logger.Error("Failed reading file from request body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError.Error(),
		})
		return
	}

	reqData := structs.RequestSubmit{
		UserID:      userID.(int64),
		ProblemID:   problemID,
		Code:        buffer,
		FileName:    "filename", // File name currently not preserved
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

	resp, contentType, status := h.submissionsHandler.Get(c, userID.(int64), submissionID)
	if status == http.StatusOK {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.Metadata.FileName))
		c.Data(status, contentType, resp.RawCode)
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
