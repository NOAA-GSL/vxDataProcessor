package api

import (
	"net/http"
	"strconv"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/gin-gonic/gin"
)

// jobServer has access to a JobStore and defines the handlers for the API server
type jobServer struct {
	store *jobstore.JobStore
}

// NewJobServer returns an initialized jobServer pointer.
//
// It can take a pointer to an existing JobStore to store Jobs in. If nil is
// passed instead of a JobStore, it will initialize an empty JobStore for you.
func NewJobServer(js *jobstore.JobStore) *jobServer {
	if js == nil {
		js = jobstore.NewJobStore()
	}

	return &jobServer{store: js}
}

// getAllJobsHandler handles requests to get all of the Jobs in the store
func (js *jobServer) getAllJobsHandler(c *gin.Context) {
	allJobs := js.store.GetAllJobs()
	c.JSON(http.StatusOK, allJobs)
}

// createJobHandler handles requests to create a new Job in the store
func (js *jobServer) createJobHandler(c *gin.Context) {
	type RequestJob struct {
		DocID string `json:"docid" binding:"required"`
	}

	var rj RequestJob
	if err := c.ShouldBindJSON(&rj); err != nil {
		c.String(http.StatusBadRequest, err.Error()) // TODO: Better error message
		return
	}

	id, err := js.store.CreateJob(rj.DocID)
	if err.Error() == "docID already exists" {
		c.String(http.StatusBadRequest, err.Error())
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, err.Error()) // TODO: Better error message
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// getJobHandler handles requests to get a specific Job in the store
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
