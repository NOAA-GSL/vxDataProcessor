package jobstore

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/exp/slices"
)

func TestNewJobStore(t *testing.T) {
	tests := []struct {
		name string
		want *JobStore
	}{
		// test cases
		{
			name: "Test New Job Store",
			want: &JobStore{
				jobs:   map[int]Job{},
				nextId: 0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NewJobStore(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewJobStore() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestJobStore_CreateJob(t *testing.T) {
	t.Run("Test creating a job", func(t *testing.T) {
		jobstore := NewJobStore()
		want := Job{
			Id:      0,
			DocHash: "foo",
			Status:  "created",
		}

		jobstore.CreateJob("foo")

		got, _ := jobstore.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.CreateJob() = %v, want %v", got, want)
		}
	})

	t.Run("Test creating a second job", func(t *testing.T) {
		jobstore := NewJobStore()
		want := Job{
			Id:      1,
			DocHash: "bar",
			Status:  "created",
		}

		jobstore.CreateJob("foo")
		jobstore.CreateJob("bar")

		got, _ := jobstore.GetJob(1)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.CreateJob() = %v, want %v", got, want)
		}
	})

	// TODO - what happens if we get the same docID submitted multiple times?
}

func TestJobStore_GetJob(t *testing.T) {
	t.Run("Test getting a nonexistant job", func(t *testing.T) {
		jobstore := NewJobStore()

		jobID := 0
		_, err := jobstore.GetJob(jobID)
		if err == nil {
			t.Error("JobStore.GetJob() Wanted an error but didn't get one")
			return
		}

		wantedErr := fmt.Sprintf("job with id=%v not found", jobID)
		if err != nil && err.Error() != wantedErr {
			t.Errorf("JobStore.GetJob() got error: '%v', want error: '%v'", err, wantedErr)
			return
		}
	})

	t.Run("Test getting a job", func(t *testing.T) {
		jobstore := NewJobStore()
		jobstore.CreateJob("foo")

		want := Job{
			Id:      0,
			DocHash: "foo",
			Status:  "created",
		}
		got, _ := jobstore.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.GetJob() = %v, want %v", got, want)
		}
	})

	t.Run("Test getting the correct job", func(t *testing.T) {
		jobstore := NewJobStore()
		jobstore.CreateJob("foo")
		jobstore.CreateJob("bar")

		want := Job{
			Id:      0,
			DocHash: "foo",
			Status:  "created",
		}
		got, _ := jobstore.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.GetJob() = %v, want %v", got, want)
		}
	})
}

func TestJobStore_GetAllJobs(t *testing.T) {
	t.Run("Test getting multiple jobs", func(t *testing.T) {
		jobstore := NewJobStore()
		jobstore.CreateJob("foo")
		jobstore.CreateJob("bar")

		want := []Job{
			{Id: 0, DocHash: "foo", Status: "created"},
			{Id: 1, DocHash: "bar", Status: "created"},
		}
		got := jobstore.GetAllJobs()
		if !slices.Contains(got, want[0]) {
			t.Errorf("JobStore.GetAllJobs() Job %v not in jobs: %v", want[0], got)
		}
		if !slices.Contains(got, want[1]) {
			t.Errorf("JobStore.GetAllJobs() Job %v not in jobs: %v", want[1], got)
		}
	})
}

func TestJobStore_updateJobStatus(t *testing.T) {
	t.Run("Set to random string", func(t *testing.T) {
		jobstore := NewJobStore()
		jobstore.CreateJob("foo")

		want := Job{Id: 0, DocHash: "foo", Status: "mystatus"}
		err := jobstore.updateJobStatus(0, "mystatus")
		if err != nil {
			t.Errorf("JobStore.updateJobStatus() got an unexpected error: %v", err.Error())
			return
		}

		got, _ := jobstore.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.updateJobStatus() = %v, want %v", got, want)
		}
	})

	t.Run("Errors as expected", func(t *testing.T) {
		jobstore := NewJobStore()
		jobstore.CreateJob("foo")

		want := "job with id=1 not found"
		err := jobstore.updateJobStatus(1, "mystatus")
		if err == nil {
			t.Error("JobStore.updateJobStatus() didn't error as expected")
			return
		}

		if err.Error() != want {
			t.Errorf("JobStore.updateJobStatus got error '%v', wanted error '%v'", err.Error(), want)
		}

	})

}
