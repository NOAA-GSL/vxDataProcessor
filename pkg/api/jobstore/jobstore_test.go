package jobstore

import (
	"reflect"
	"sync"
	"testing"
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
	type fields struct {
		lock   sync.Mutex
		jobs   map[int]Job
		nextId int
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// test cases
		// TODO - how can we check the hash is as desired? I'm not sure these tests are all that useful.
		{
			name:   "Test Creating Job",
			fields: fields{lock: sync.Mutex{}, jobs: map[int]Job{}, nextId: 0},
			args:   args{hash: "myhash"},
			want:   0,
		},
		{
			name: "Test Creating Second Job",
			fields: fields{
				lock: sync.Mutex{},
				jobs: map[int]Job{
					0: {Id: 0, DocHash: "myhash", Status: "created"},
				},
				nextId: 1},
			args: args{hash: "myhash"},
			want: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			js := &JobStore{
				lock:   test.fields.lock, // TODO - avoid copying a Mutex. Pointers and sync.Locker could be interesting.
				jobs:   test.fields.jobs,
				nextId: test.fields.nextId,
			}
			if got := js.CreateJob(test.args.hash); got != test.want {
				t.Errorf("JobStore.CreateJob() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestJobStore_GetJob(t *testing.T) {
	type fields struct {
		lock   sync.Mutex
		jobs   map[int]Job
		nextId int
	}
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Job
		wantErr bool
	}{
		// test cases
		{
			name: "Test Getting Job",
			fields: fields{
				lock: sync.Mutex{},
				jobs: map[int]Job{
					0: {Id: 0, DocHash: "myhash", Status: "created"},
				},
				nextId: 0},
			args:    args{id: 0},
			want:    Job{Id: 0, DocHash: "myhash", Status: "created"},
			wantErr: false,
		},
		{
			name: "Test Getting Multiple Jobs",
			fields: fields{
				lock: sync.Mutex{},
				jobs: map[int]Job{
					0: {Id: 0, DocHash: "myhash", Status: "created"},
					1: {Id: 1, DocHash: "myhash1", Status: "created"},
				},
				nextId: 0},
			want:    Job{Id: 0, DocHash: "myhash", Status: "created"},
			wantErr: false,
		},
		{
			name:    "Test Getting Nonexistant Job",
			fields:  fields{lock: sync.Mutex{}, jobs: map[int]Job{}, nextId: 0},
			args:    args{id: 0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := &JobStore{
				lock:   tt.fields.lock,
				jobs:   tt.fields.jobs,
				nextId: tt.fields.nextId,
			}
			got, err := js.GetJob(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("JobStore.GetJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JobStore.GetJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobStore_GetAllJobs(t *testing.T) {
	type fields struct {
		lock   sync.Mutex
		jobs   map[int]Job
		nextId int
	}
	tests := []struct {
		name   string
		fields fields
		want   []Job
	}{
		{
			name: "Test Getting Multiple Jobs",
			fields: fields{
				lock: sync.Mutex{},
				jobs: map[int]Job{
					0: {Id: 0, DocHash: "myhash", Status: "created"},
					1: {Id: 1, DocHash: "myhash1", Status: "created"},
				},
				nextId: 0},
			want: []Job{
				{Id: 0, DocHash: "myhash", Status: "created"},
				{Id: 1, DocHash: "myhash1", Status: "created"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := &JobStore{
				lock:   tt.fields.lock,
				jobs:   tt.fields.jobs,
				nextId: tt.fields.nextId,
			}
			if got := js.GetAllJobs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JobStore.GetAllJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobStore_updateJobStatus(t *testing.T) {
	type fields struct {
		lock   sync.Mutex
		jobs   map[int]Job
		nextId int
	}
	type args struct {
		id     int
		status string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Assert that the status is updated as expected
		{
			name: "Set to random string",
			fields: fields{
				lock: sync.Mutex{},
				jobs: map[int]Job{
					0: {Id: 0, DocHash: "myhash", Status: "created"},
					1: {Id: 1, DocHash: "myhash1", Status: "created"},
				},
				nextId: 0},
			args:    args{0, "foo"},
			wantErr: false,
		},
		{
			name: "Errors as expected",
			fields: fields{
				lock: sync.Mutex{},
				jobs: map[int]Job{
					0: {Id: 0, DocHash: "myhash", Status: "created"},
					1: {Id: 1, DocHash: "myhash1", Status: "created"},
				},
				nextId: 0},
			args:    args{3, "foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := &JobStore{
				lock:   tt.fields.lock,
				jobs:   tt.fields.jobs,
				nextId: tt.fields.nextId,
			}
			if err := js.updateJobStatus(tt.args.id, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("JobStore.updateJobStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
