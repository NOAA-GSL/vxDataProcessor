package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingEndpoint(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"message":"pong"}`, w.Body.String())
}

func TestJobsEndpoint(t *testing.T) {
	t.Run("Test Creating a Job", func(t *testing.T) {
		router := SetupRouter()

		// Setup
		w := httptest.NewRecorder()
		var jsonStr = []byte(`{"docid": "myid1"}`)

		// Test
		req, _ := http.NewRequest("POST", "/jobs/", bytes.NewBuffer(jsonStr))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"id":0}`, w.Body.String())
	})

	t.Run("Test Getting All Jobs", func(t *testing.T) {
		router := SetupRouter()

		// Setup
		// TODO - is there a better way to insert state?
		w := httptest.NewRecorder()
		var jsonStr = []byte(`{"docid": "myid1"}`)
		req, _ := http.NewRequest("POST", "/jobs/", bytes.NewBuffer(jsonStr))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Test
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/jobs/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `[{"id":0,"docid":"myid1","status":"created"}]`, w.Body.String())
	})
}

func TestJobsIDEndpoint(t *testing.T) {
	router := SetupRouter()

	// Setup
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{"docid": "myid1"}`)
	req, _ := http.NewRequest("POST", "/jobs/", bytes.NewBuffer(jsonStr))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/jobs/0", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"id":0,"docid":"myid1","status":"created"}`, w.Body.String())
}
