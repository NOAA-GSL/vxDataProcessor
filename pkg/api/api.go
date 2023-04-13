// Package api implements a Gin API server and handlers for the data processor.
package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter defines the routes the API server will respond to along with
// their handlers
func SetupRouter(js *jobstore.JobStore) *gin.Engine {
	router := gin.Default()
	server := newJobServer(js)
	router.Use(prometheusMiddleware()) // attach our Prometheus middleware

	router.POST("/jobs/", server.createJobHandler)
	router.GET("/jobs/", server.getAllJobsHandler)
	router.GET("/jobs/:id", server.getJobHandler)
	router.GET(defaultMetricPath, gin.WrapH(promhttp.Handler())) // expose Prometheus metrics

	// healthcheck
	router.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return router
}

// Processor is an interface used to inject calculation functions into the Worker
// processor is intended to encapsulate the manager.manager struct
type Processor interface {
	Run() error
}

// Worker receives jobs on a channel, processes them, and reports the status on a return channel
func Worker(id int, getProcessor func(string, string) (Processor, error), jobs <-chan jobstore.Job, status chan<- jobstore.Job) {
	for {
		job := <-jobs // block until we get a job
		fmt.Println("Worker", id, "started docID", job.DocID)
		job.Status = jobstore.StatusProcessing
		status <- job

		// Do work
		start := time.Now()
		fmt.Println("Worker", id, "processing docID", job.DocID)

		docType := strings.Split(job.DocID, ":")[0]
		mgr, err := getProcessor(docType, job.DocID)
		if err != nil {
			job.Status = jobstore.StatusFailed
			status <- job
			return
		}

		err = mgr.Run()
		duration := time.Since(start).Seconds()
		calculationDuration.WithLabelValues(job.DocID).Observe(duration)
		if err != nil {
			job.Status = jobstore.StatusFailed
			status <- job
			return
		}

		// report status
		job.Status = jobstore.StatusCompleted
		status <- job
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

// StatusUpdater receives jobs to update in the given jobstore. It will block if the channel is empty.
func StatusUpdater(statusChan <-chan jobstore.Job, js *jobstore.JobStore) {
	for {
		job := <-statusChan
		err := js.UpdateJobStatus(job.ID, job.Status)
		if err != nil {
			fmt.Printf("Error - StatusUpdater: %v\n", err.Error())
		}
	}
}
