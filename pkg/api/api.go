package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Start() {
	router := gin.Default()
	server := NewJobServer()

	router.POST("/jobs/", server.createJobHandler)
	router.GET("/jobs/", server.getAllJobsHandler)
	router.GET("/jobs/:id", server.getJobHandler)

	// healthcheck
	router.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.Run() // listen and serve on 0.0.0.0:8080
}
