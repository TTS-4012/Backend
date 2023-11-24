package api

import (
	"net/http"
	"ocontest/pkg"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *handlers) UploadSubmission(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "UploadSubmission")

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

	//TODO: get submission id instead of "1"
	objectName := strconv.FormatInt(userID.(int64), 10) + "-" + "1"

	resp, status := h.submissionsHandler.UploadFile(c, file, objectName)
	c.JSON(status, resp)
}

func (h *handlers) DownloadSubmission(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "DownloadSubmission")

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
