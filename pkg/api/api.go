// Package api implements a Gin API server and handlers for the data processor.
package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/gin-gonic/gin"
)

// SetupRouter defines the routes the API server will respond to along with
// their handlers
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

// Processor is an interface used to inject calculation functions into the Worker
type Processor interface {
	Run(string) error
}

// TestProcess Implements the Calculator interface with some hooks for triggering testing behavior
type TestProcess struct {
	// FIXME - this will be moved to the _test file once we have a real processor to use
	lock         sync.Mutex
	DocID        string
	Processed    bool
	TriggerError bool
}

// Run is a dummy method for testing that satisfies the Processor interface
func (tc *TestProcess) Run(str string) error {
	// FIXME - this will be moved to the _test file once we have a real processor to use
	tc.lock.Lock()
	defer tc.lock.Unlock()
	if tc.TriggerError {
		return fmt.Errorf("Unable to process %v", tc.DocID)
	}
	tc.DocID = str
	fmt.Println("Processed", tc.DocID)
	tc.Processed = true
	return nil
}

// Worker receives jobs on a channel, processes them, and reports the status on a return channel
func Worker(id int, proc Processor, jobs <-chan jobstore.Job, status chan<- string) {
	for {
		job := <-jobs // block until we get a job
		fmt.Println("Worker", id, "started docID", job.DocID)
		// status <- "processing" // We'll need a way to associate these with a job

		// Do work
		fmt.Println("Worker", id, "processing docID", job.DocID)
		err := proc.Run(job.DocID)
		if err != nil {
			status <- fmt.Sprintf("Unable to process %v", job.DocID)
		}

		// report status
		status <- fmt.Sprintf("Finished %v", job.DocID)
		fmt.Println("Worker", id, "finished docID", job.DocID)
	}
}

// Dispatch pulls jobs out of the given jobstore in order and places them in a channel. It will block once the channel is full.
func Dispatch(jobChan chan<- jobstore.Job, js *jobstore.JobStore) {
	for {
		n := 2 // number of jobs to pull off the queue
		jobs, err := js.GetJobsToProcess(n)
		if err != nil {
			if err.Error() == "No unprocessed jobs available" {
				continue
			}
			panic(err)
		}

		for _, job := range jobs {
			jobChan <- job
			fmt.Printf("Dispatched item #%v\n", job)
		}
	}
}

