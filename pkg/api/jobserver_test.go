package api

import (
	"bytes"
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
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		js := NewJobServer()

		js.store.CreateJob("foo")
		js.store.CreateJob("bar")
		js.getAllJobsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `[{"id":0,"docid":"foo","status":"created"},{"id":1,"docid":"bar","status":"created"}]`, w.Body.String())
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
