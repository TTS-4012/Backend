package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"judge-code/pkg"
	"net/http"
)

func main() {
	pkg.InitConf()
	c := pkg.Conf

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	addr := fmt.Sprintf("0.0.0.0:%d", c.Port)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
