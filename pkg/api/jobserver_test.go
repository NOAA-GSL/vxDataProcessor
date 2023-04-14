package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type errResponse struct {
	Code    int    `json:"code" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func TestNewJobServer(t *testing.T) {
	filledJS := jobstore.NewJobStore()
	_, _ = filledJS.CreateJob("foo")

	type args struct {
		js *jobstore.JobStore
	}
	tests := []struct {
		name string
		args args
		want *jobServer
	}{
		{
			name: "Test New Empty JobServer",
			args: args{nil},
			want: &jobServer{
				store: jobstore.NewJobStore(),
			},
		},
		{
			name: "Test New Filled JobServer",
			args: args{filledJS},
			want: &jobServer{
				store: filledJS,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newJobServer(tt.args.js); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJobServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jobServer_getAllJobsHandler(t *testing.T) {
	t.Run("Test getting an empty jobstore", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		newJobServer(nil).getAllJobsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `[]`, w.Body.String())
	})

	t.Run("Test getting a jobstore", func(t *testing.T) {
		want := []jobstore.Job{
			{ID: 0, DocID: "foo", Status: jobstore.StatusCreated},
			{ID: 1, DocID: "bar", Status: jobstore.StatusCreated},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		js := newJobServer(nil)

		_, _ = js.store.CreateJob("foo")
		_, _ = js.store.CreateJob("bar")

		js.getAllJobsHandler(c)
		got := []jobstore.Job{}
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal("Issue unmarshalling JSON response")
		}

		assert.Equal(t, http.StatusOK, w.Code)
		assert.ElementsMatch(t, want, got)
	})
}

func Test_jobServer_createJobHandler(t *testing.T) {
	t.Run("Test a bad job submission", func(t *testing.T) {
		want := errResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON - expecting a 'docid' key",
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonStr := []byte(`{"random": "SC:json"}`)
		c.Request, _ = http.NewRequest(http.MethodPost, "/jobs/", bytes.NewBuffer(jsonStr))

		js := newJobServer(nil)

		js.createJobHandler(c)

		got := errResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal("Issue unmarshalling JSON response")
		}

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, got, want)
	})

	t.Run("Test a duplicate job submission", func(t *testing.T) {
		want := errResponse{
			Code:    http.StatusBadRequest,
			Message: "That docid already exists",
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonStr := []byte(`{"docid": "SC:json"}`)
		c.Request, _ = http.NewRequest(http.MethodPost, "/jobs/", bytes.NewBuffer(jsonStr))

		store := jobstore.NewJobStore()
		_, _ = store.CreateJob("SC:json")
		js := newJobServer(store)

		js.createJobHandler(c)

		got := errResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal("Issue unmarshalling JSON response")
		}

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, want, got)
	})
}
