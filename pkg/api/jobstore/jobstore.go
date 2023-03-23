// Package jobstore implements an in-memory store of all jobs submitted to the
// api server.
//
// This could theoretically be better represented as a queue. The Job struct
// is a representation of a single job while the JobStore represents a
// collection of jobs. At the moment there is no persistence to disk so if the
// program is stopped and restarted all Jobs will be lost. This could be
// handled more gracefully. Additionally, Jobs are never removed from the
// JobStore. In theory, we have an int64's worth (2^63) of Job's before
// we run into issues.
package jobstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

type JobStatus int

const (
	StatusCreated JobStatus = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
)

// String supports pretty-printing JobStatuses
func (js JobStatus) String() string {
	return []string{"created", "processing", "completed", "failed"}[js]
}

// toString is an internal helper function for marshalling to JSON
var toString = map[JobStatus]string{
	StatusCreated:    "created",
	StatusProcessing: "processing",
	StatusCompleted:  "completed",
	StatusFailed:     "failed",
}

// toID is an internal helper function for unmarshalling from JSON
var toID = map[string]JobStatus{
	"created":    StatusCreated,
	"processing": StatusProcessing,
	"completed":  StatusCompleted,
	"failed":     StatusFailed,
}

// MarshalJSON supports writing the iota to JSON as a string
func (js JobStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[js])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON supports converting JSON strings to our iota
func (js *JobStatus) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*js = toID[j]
	return nil
}

// Job represents a single job in the JobStore
type Job struct {
	ID     int       `json:"id"`
	DocID  string    `json:"docid"`
	Status JobStatus `json:"status"` // Zero Value is "StatusCreated"
}

// FIXME - we'll want to handle removing Jobs from the JobStore so we don't
// need to worry about handling int overflows if this has been running for a
// while. A queue might be a better implementation.

// JobStore contains a map of all Jobs submitted for processing.
// It is intended to be thread-safe.
type JobStore struct {
	lock            sync.RWMutex // lock for modifying jobs & nextID
	processLock     sync.Mutex   // lock for updating nextIDToProcess
	jobs            map[int]Job
	nextID          int
	nextIDToProcess int
}

func NewJobStore() *JobStore {
	js := &JobStore{}
	js.jobs = make(map[int]Job)
	return js
}

// CreateJob creates a new job in the store and returns the int key to access it
func (js *JobStore) CreateJob(docID string) (int, error) {
	js.lock.Lock()
	defer js.lock.Unlock()

	// FIXME: Test for and dissallow duplicate docID values
	if docID == "" {
		return 0, fmt.Errorf("expected a non-empty docID")
	}

	job := Job{
		ID:     js.nextID,
		DocID:  docID,
		Status: StatusCreated,
	}

	js.jobs[js.nextID] = job
	js.nextID++
	return job.ID, nil
}

// GetJob retrieves a job from the store, by id. If no such id exists, an
// error is returned.
func (js *JobStore) GetJob(id int) (Job, error) {
	js.lock.RLock()
	defer js.lock.RUnlock()

	j, ok := js.jobs[id]
	if ok {
		return j, nil
	} else {
		return Job{}, fmt.Errorf("job with id=%d not found", id)
	}
}

// GetAllJobs returns all the jobs in the store, in arbitrary order.
func (js *JobStore) GetAllJobs() []Job {
	js.lock.RLock()
	defer js.lock.RUnlock()

	allJobs := make([]Job, 0, len(js.jobs))
	for _, job := range js.jobs {
		allJobs = append(allJobs, job)
	}
	return allJobs
}

// GetJobsToProcess returns up to the numJobs number of jobs that haven't been processed
func (js *JobStore) GetJobsToProcess(numJobs int) ([]Job, error) {
	js.processLock.Lock()
	defer js.processLock.Unlock()

	jobsToProcess := []Job{}
	for i := js.nextIDToProcess; i < js.nextIDToProcess+numJobs; i++ {
		j, err := js.GetJob(i)
		if err != nil {
			// No job with that ID, return
			if len(jobsToProcess) == 0 {
				return []Job{}, fmt.Errorf("No unprocessed jobs available")
			} else {
				// Return what we have
				js.nextIDToProcess = js.nextIDToProcess + len(jobsToProcess)
				return jobsToProcess, nil
			}
		}
		jobsToProcess = append(jobsToProcess, j)
	}
	js.nextIDToProcess = js.nextIDToProcess + len(jobsToProcess)
	return jobsToProcess, nil
}

// updateJobStatus changes the status of the Job to the specified JobStatus.
//
// It returns an error if the Job doesn't exist or if it's already been set to finished.
func (js *JobStore) UpdateJobStatus(id int, status JobStatus) error {
	js.lock.Lock()
	defer js.lock.Unlock()

	job, ok := js.jobs[id]
	if !ok {
		return fmt.Errorf("job with id=%d not found", id)
	}

	if job.Status == StatusCompleted {
		return fmt.Errorf("Job already marked as completed.")
	}
	job.Status = status

	js.jobs[id] = job
	return nil
}
