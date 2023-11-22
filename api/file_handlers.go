package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handlers) UploadFile(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *handlers) DownloadFile(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
