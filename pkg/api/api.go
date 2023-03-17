package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
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

func Worker(id int, jobs <-chan jobstore.Job, status chan<- string) {
	for {
		job := <-jobs // block until we get a job
		fmt.Println("worker", id, "started  docID", job.DocID)
		// status <- "processing" // We'll need a way to associate these with a job

		// Do work
		fmt.Println("Worker", id, "processing docID", job.DocID)
		time.Sleep(time.Second)

		// report status
		// status <- "finished"
		fmt.Println("worker", id, "finished docID", job.DocID)
	}
}

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

