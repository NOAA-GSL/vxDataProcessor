package jobstore

import (
	"fmt"
	"sync"
)

// Contains map of all jobs

type Job struct {
	ID     int    `json:"id"`
	DocID  string `json:"docid"`
	Status string `json:"status"`
}

type JobStore struct {
	lock   sync.Mutex
	jobs   map[int]Job
	nextId int
}

func NewJobStore() *JobStore {
	js := &JobStore{}
	js.jobs = make(map[int]Job)
	js.nextId = 0
	return js
}

// CreateJob creates a new job in the store.
func (js *JobStore) CreateJob(docID string) (int, error) {
	js.lock.Lock()
	defer js.lock.Unlock()

	if docID == "" {
		return 0, fmt.Errorf("expected a non-empty docID")
	}

	job := Job{
		ID:     js.nextId,
		DocID:  docID,
		Status: "created", // what statuses do we want? Created, Processing, Finished, Failed?
	}

	js.jobs[js.nextId] = job
	js.nextId++
	return job.ID, nil
}

// GetJob retrieves a job from the store, by id. If no such id exists, an
// error is returned.
func (js *JobStore) GetJob(id int) (Job, error) {
	js.lock.Lock()
	defer js.lock.Unlock()

	j, ok := js.jobs[id]
	if ok {
		return j, nil
	} else {
		return Job{}, fmt.Errorf("job with id=%d not found", id)
	}
}

// GetAllJobs returns all the jobs in the store, in arbitrary order.
func (js *JobStore) GetAllJobs() []Job {
	js.lock.Lock()
	defer js.lock.Unlock()

	allJobs := make([]Job, 0, len(js.jobs))
	for _, job := range js.jobs {
		allJobs = append(allJobs, job)
	}
	return allJobs
}

// UpdateJobStatus changes the status of the job to the specified string
func (js *JobStore) updateJobStatus(id int, status string) error {
	js.lock.Lock()
	defer js.lock.Unlock()

	j, ok := js.jobs[id]
	if ok {
		j.Status = status
		js.jobs[id] = j
		return nil
	} else {
		return fmt.Errorf("job with id=%d not found", id)
	}
}

// Convenience function to set the job status to processing
func (js *JobStore) SetJobStatusProcessing(id int) error {
	return js.updateJobStatus(id, "processing")
}

// Convenience function to set the job status to completed
func (js *JobStore) SetJobStatusCompleted(id int) error {
	return js.updateJobStatus(id, "completed")
}

// Convenience function to set the job status to failed
func (js *JobStore) SetJobStatusFailed(id int) error {
	return js.updateJobStatus(id, "failed")
}
