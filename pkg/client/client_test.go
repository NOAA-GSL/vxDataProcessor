package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_EmptyPOSTRequest(t *testing.T) {
	expectedDocID := "foo"
	expectedURL := fmt.Sprintf("/refreshScorecard/%v", expectedDocID)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != expectedURL {
			t.Errorf("Expected request to %s, got %s", expectedURL, r.URL.Path)
		}
		if r.ContentLength != 0 {
			t.Errorf("Expected empty request body, got length %d", r.ContentLength)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	err := UpdateMATS(svr.URL, expectedDocID)
	if err != nil {
		t.Errorf("Unexpected error")
	}
}

func TestClient_StatusNotFound(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer svr.Close()

	URL := fmt.Sprintf("%s/refreshScorecard/foo", svr.URL)
	want := fmt.Sprintf("client: got response code 404 from %s", URL)
	err := UpdateMATS(svr.URL, "foo")
	if err.Error() != want {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestClient_WrongURL(t *testing.T) {
	err := UpdateMATS("http://localhost:1000", "foo")
	if err == nil {
		t.Errorf("Expected an Error to occur")
	}
	if !strings.HasPrefix(err.Error(), "client: error making http request:") {
		t.Errorf("Unexpected error: %v", err)
	}
}
