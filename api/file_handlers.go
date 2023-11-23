package api

import (
	"net/http"
	"ocontest/pkg"

	"github.com/gin-gonic/gin"
)

func (h *handlers) UploadFile(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "UploadFile")

	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to read request body file", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.filesHandler.UploadFile(c, file)
	c.JSON(status, resp)
}

func (h *handlers) DownloadFile(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "DownloadFile")
	objectName := c.Param("name")
	if objectName == "" {
		logger.Error("Failed to read object name")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.filesHandler.DownloadFile(c, objectName)
	c.Header("Content-Disposition", resp.ContentDisposition)
	c.Data(status, resp.ContentType, resp.Data)
}
