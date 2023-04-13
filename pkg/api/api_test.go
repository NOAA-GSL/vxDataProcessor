package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/stretchr/testify/assert"
)

// TestProcess Implements the Processor interface with some hooks for triggering testing behavior
type TestProcess struct {
	lock         sync.Mutex
	DocID        string
	Processed    bool
	TriggerError bool
}

// Run is a dummy method for testing that satisfies the Processor interface
func (tp *TestProcess) Run() error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	if tp.TriggerError {
		return fmt.Errorf("TestProcess - Unable to process %v", tp.DocID)
	}
	fmt.Println("TestProcess - Processed", tp.DocID)
	tp.Processed = true
	return nil
}

func ProcessorFactoryMock(docID string) (Processor, error) {
	documentType := strings.Split(docID, ":")[0]
	switch documentType {
	case "SC":
		return &TestProcess{DocID: docID}, nil
	case "Err":
		return &TestProcess{DocID: docID, TriggerError: true}, nil
	default:
		return nil, fmt.Errorf("Unknown processor type")
	}
}

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
			DocID:  "SC:foo",
			Status: jobstore.StatusCreated,
		}
		go Worker(1, ProcessorFactoryMock, jobs, status)
		jobs <- job

		want := []jobstore.Job{
			{ID: 1, DocID: "SC:foo", Status: jobstore.StatusProcessing},
			{ID: 1, DocID: "SC:foo", Status: jobstore.StatusCompleted},
		}
		for {
			select {
			case got1 := <-status:
				got2 := <-status // ignore the processing status
				assert.Equal(t, want[0], got1)
				assert.Equal(t, want[1], got2)
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
			DocID:  "Err:foo",
			Status: jobstore.StatusCreated,
		}
		go Worker(2, ProcessorFactoryMock, jobs, status)
		jobs <- job

		want := []jobstore.Job{
			{ID: 1, DocID: "Err:foo", Status: jobstore.StatusProcessing},
			{ID: 1, DocID: "Err:foo", Status: jobstore.StatusFailed},
		}
		for {
			select {
			case got1 := <-status:
				got2 := <-status // ignore the processing status
				assert.Equal(t, want[0], got1)
				assert.Equal(t, want[1], got2)
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
