package main

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
)

func main() {
	jobs := make(chan jobstore.Job)
	status := make(chan string) // this channel needs a status and jobID
	js := jobstore.NewJobStore()

	// setup dispatcher to send jobs from queue to worker pool
	// https://webdevstation.com/posts/simple-queue-implementation-in-golang/
	go api.Dispatch(jobs, js)

	// setup worker pool to recieve jobs to pass on to process and to send status updates
	for w := 1; w <= 5; w++ {
		go api.Worker(w, jobs, status)
	}

	router := api.SetupRouter(js)

	err := router.Run(":8080") // listen and serve on 0.0.0.0:8080
	if err != nil {
		panic(err)
	}
}
