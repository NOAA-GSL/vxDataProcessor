package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
func newJobServer(js *jobstore.JobStore) *jobServer {
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

// isValidDocType tests the scorecard docType passed in to make sure it's valid
func isValidDocType(docType string) bool {
	switch docType {
	case "SC":
		return true
	default:
		return false
	}
}

// createJobHandler handles requests to create a new Job in the store
func (js *jobServer) createJobHandler(c *gin.Context) {
	type RequestJob struct {
		DocID string `json:"docid" binding:"required"`
	}

	var rj RequestJob
	if err := c.ShouldBindJSON(&rj); err != nil {
		// TODO - Use error handling middleware & c.Error
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid JSON - expecting a 'docid' key",
		})
		return
	}

	// test the DocID is a valid type
	docType := strings.Split(rj.DocID, ":")[0]
	if !isValidDocType(docType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Invalid scorecard document type %v", docType),
		})
		return
	}

	id, err := js.store.CreateJob(rj.DocID)
	if err != nil {
		if err.Error() == "docID already exists" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "That docid already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// getJobHandler handles requests to get a specific Job in the store
func (js *jobServer) getJobHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Unable to parse job id \"%v\", an int is required", c.Params.ByName("id")),
		})
		return
	}

	job, err := js.store.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, job)
}
