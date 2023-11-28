package api

import (
	"net/http"
	"ocontest/pkg"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *handlers) Submit(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "Submit")

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError,
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to read request body file", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	h.submissionsHandler.Submit(c, userID)
}

func (h *handlers) GetSubmission(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "GetSubmission")

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError,
		})
		return
	}

	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id, id should be an integer",
		})
		return
	}

	objectName := strconv.FormatInt(userID.(int64), 10) + "-" + strconv.FormatInt(submissionID, 10)

	resp, status := h.submissionsHandler.DownloadFile(c, objectName)
	c.Header("Content-Disposition", resp.ContentDisposition)
	c.Data(status, resp.ContentType, resp.Data)
}
