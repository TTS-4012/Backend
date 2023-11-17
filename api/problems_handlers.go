package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/pkg"
	"ocontest/pkg/structs"
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
	c.JSON(status, resp)
}

func (h *handlers) GetProblem(c *gin.Context) {
	h.problemsHandler.GetProblem(c, 0)
}

func (h *handlers) ListProblems(c *gin.Context) {
	h.problemsHandler.ListProblem(c, structs.RequestListProblems{})
}
