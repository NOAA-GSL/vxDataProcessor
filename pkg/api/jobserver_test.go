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

func TestNewJobServer2(t *testing.T) {
	filledJS := jobstore.NewJobStore()
	filledJS.CreateJob("foo")

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
			if got := NewJobServer(tt.args.js); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJobServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jobServer_getAllJobsHandler(t *testing.T) {
	t.Run("Test getting an empty jobstore", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		NewJobServer(nil).getAllJobsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `[]`, w.Body.String())
	})

	t.Run("Test getting a jobstore", func(t *testing.T) {
		want := []jobstore.Job{
			{ID: 0, DocID: "foo", Status: "created"},
			{ID: 1, DocID: "bar", Status: "created"},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		js := NewJobServer(nil)

		_, _ = js.store.CreateJob("foo")
		_, _ = js.store.CreateJob("bar")

		js.getAllJobsHandler(c)
		got := []jobstore.Job{}
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Error("Issue unmarshalling JSON response")
		}

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

		js := NewJobServer(nil)

		js.createJobHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
