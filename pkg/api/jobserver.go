package api

import (
	"net/http"
	"strconv"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/gin-gonic/gin"
)

// Contains store of current jobs and handlers for routes

type jobServer struct {
	store *jobstore.JobStore
}

func NewJobServer() *jobServer {
	store := jobstore.NewJobStore()
	return &jobServer{store: store}
}

// Handles requests to get all Jobs in the store
func (js *jobServer) getAllJobsHandler(c *gin.Context) {
	allJobs := js.store.GetAllJobs()
	c.JSON(http.StatusOK, allJobs)
}

// Handles requests to create a new Job in the store
func (js *jobServer) createJobHandler(c *gin.Context) {
	type RequestJob struct {
		DocID string `json:"docid"`
	}

	var rj RequestJob
	if err := c.ShouldBindJSON(&rj); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	id := js.store.CreateJob(rj.DocID)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Handles requests to get a specific Job in the store
func (js *jobServer) getJobHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	job, err := js.store.GetJob(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, job)
}
