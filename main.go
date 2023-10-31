package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"ocontest/api"
	"ocontest/internal/db"
	"ocontest/internal/jwt"
	"ocontest/internal/oc/auth"
	"ocontest/pkg"
	"ocontest/pkg/configs"
)

func main() {
	configs.InitConf()
	c := configs.Conf
	pkg.InitLog(c.Log)
	pkg.Log.Info("config and log modules initialized")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	ctx := context.Background()
	// connecting to dependencies
	jwtHandler := jwt.NewGenerator(c.JWT)

	dbConn, err := db.NewConnectionPool(ctx, c.Postgres)
	if err != nil {
		log.Fatal("error on connecting to db", err)
	}

	// make repo
	authRepo, err := db.NewAuthRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating auth repo", err)
	}

	// initiating module handlers
	authHandler := auth.NewAuthHandler(authRepo, jwtHandler)

	// starting http server
	api.AddRoutes(r, authHandler)

	addr := fmt.Sprintf("0.0.0.0:%d", c.Port)
	pkg.Log.Info("Running on address: ", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
