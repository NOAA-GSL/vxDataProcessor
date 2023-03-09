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

func TestNewJobServer(t *testing.T) {
	tests := []struct {
		name string
		want *jobServer
	}{
		{
			name: "Test New JobServer",
			want: &jobServer{
				store: jobstore.NewJobStore(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewJobServer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJobServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jobServer_getAllJobsHandler(t *testing.T) {
	t.Run("Test getting an empty jobstore", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		NewJobServer().getAllJobsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `[]`, w.Body.String())
	})

	t.Run("Test getting jobstore", func(t *testing.T) {
		want := []jobstore.Job{
			{ID: 0, DocID: "foo", Status: "created"},
			{ID: 1, DocID: "bar", Status: "created"},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		js := NewJobServer()

		_, _ = js.store.CreateJob("foo")
		_, _ = js.store.CreateJob("bar")

		js.getAllJobsHandler(c)
		got := []jobstore.Job{}
		json.Unmarshal(w.Body.Bytes(), &got)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.ElementsMatch(t, want, got)
	})
}

func Test_jobServer_createJobHandler(t *testing.T) {
	t.Run("Test a bad job submission", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		var jsonStr = []byte(`{"random": "json"}`)
		c.Request, _ = http.NewRequest("POST", "/jobs/", bytes.NewBuffer(jsonStr))

		js := NewJobServer()

		js.createJobHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
