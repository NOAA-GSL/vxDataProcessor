package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/stretchr/testify/assert"
)

func TestPingEndpoint(t *testing.T) {
	router := SetupRouter(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"message":"pong"}`, w.Body.String())
}

func TestJobsEndpoint(t *testing.T) {
	t.Run("Test Creating a Job", func(t *testing.T) {
		router := SetupRouter(nil)

		// Setup
		w := httptest.NewRecorder()
		jsonStr := []byte(`{"docid": "myid1"}`)

		// Test
		req, _ := http.NewRequest(http.MethodPost, "/jobs/", bytes.NewBuffer(jsonStr))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"id":0}`, w.Body.String())
	})

	t.Run("Test Getting All Jobs", func(t *testing.T) {
		router := SetupRouter(nil)

		// Setup
		// TODO - is there a better way to insert state?
		w := httptest.NewRecorder()
		jsonStr := []byte(`{"docid": "myid1"}`)
		req, _ := http.NewRequest(http.MethodPost, "/jobs/", bytes.NewBuffer(jsonStr))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Test
		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "/jobs/", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `[{"id":0,"docid":"myid1","status":"created"}]`, w.Body.String())
	})
}

func TestJobsIDEndpoint(t *testing.T) {
	router := SetupRouter(nil)

	// Setup
	w := httptest.NewRecorder()
	jsonStr := []byte(`{"docid": "myid1"}`)
	req, _ := http.NewRequest(http.MethodPost, "/jobs/", bytes.NewBuffer(jsonStr))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/jobs/0", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"id":0,"docid":"myid1","status":"created"}`, w.Body.String())
}

func TestWorker(t *testing.T) {
	t.Run("Test that jobs are sent to the proc function", func(t *testing.T) {
		jobs := make(chan jobstore.Job)
		status := make(chan jobstore.Job)
		job := jobstore.Job{
			ID:     1,
			DocID:  "foo",
			Status: jobstore.StatusCreated,
		}
		proc := &TestProcess{}
		go Worker(1, proc, jobs, status)
		jobs <- job

		want := []jobstore.Job{
			{ID: 1, DocID: "foo", Status: jobstore.StatusProcessing},
			{ID: 1, DocID: "foo", Status: jobstore.StatusCompleted},
		}
		for {
			select {
			case got1 := <-status:
				got2 := <-status // ignore the processing status
				proc.lock.Lock()
				assert.Equal(t, want[0], got1)
				assert.Equal(t, "foo", proc.DocID)
				assert.Equal(t, true, proc.Processed)
				assert.Equal(t, want[1], got2)
				proc.lock.Unlock()
				return
			default:
				continue
			}
		}
	})

	t.Run("Test that we handle errors", func(t *testing.T) {
		jobs := make(chan jobstore.Job)
		status := make(chan jobstore.Job)
		job := jobstore.Job{
			ID:     1,
			DocID:  "foo",
			Status: jobstore.StatusCreated,
		}
		proc := &TestProcess{
			TriggerError: true,
		}
		go Worker(2, proc, jobs, status)
		jobs <- job

		want := []jobstore.Job{
			{ID: 1, DocID: "foo", Status: jobstore.StatusProcessing},
			{ID: 1, DocID: "foo", Status: jobstore.StatusFailed},
		}
		for {
			select {
			case got1 := <-status:
				got2 := <-status // ignore the processing status
				assert.Equal(t, want[0], got1)
				assert.Equal(t, want[1], got2)
				assert.Equal(t, false, proc.Processed)
				return
			default:
				continue
			}
		}
	})
}

func TestDispatch(t *testing.T) {
	t.Run("Test that jobs are dispatched", func(t *testing.T) {
		jobs := make(chan jobstore.Job)
		js := jobstore.NewJobStore()
		_, _ = js.CreateJob("foo1")

		go Dispatch(jobs, js)

		for {
			select {
			case j := <-jobs:
				assert.Equal(t, "foo1", j.DocID)
				return
			default:
				continue
			}
		}
	})
}
