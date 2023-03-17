package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(js *jobstore.JobStore) *gin.Engine {
	router := gin.Default()
	server := NewJobServer(js)

	router.POST("/jobs/", server.createJobHandler)
	router.GET("/jobs/", server.getAllJobsHandler)
	router.GET("/jobs/:id", server.getJobHandler)

	// healthcheck
	router.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return router
}
