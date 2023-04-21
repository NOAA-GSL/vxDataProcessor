package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	err := NotifyScorecard(svr.URL, expectedDocID)
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
	err := NotifyScorecard(svr.URL, "foo")
	if err.Error() != want {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestClient_WrongURL(t *testing.T) {
	err := NotifyScorecard("http://localhost:1000", "foo")
	if err == nil {
		t.Errorf("Expected an Error to occur")
	}
	if !strings.HasPrefix(err.Error(), "client: error making http request:") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestClient_PathEscaped(t *testing.T) {
	DocID := "SC:anonymous--submitted:20230322220711--2block:0:02/19/2023_20_00_-_03/21/2023_20_00"
	escapedDocID := url.PathEscape(DocID)
	expectedURL := fmt.Sprintf("/refreshScorecard/%v", escapedDocID)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Path automatically unescapes URL-encoded paths so we need to use the RawPath here.
		if r.URL.RawPath != expectedURL {
			t.Errorf("Expected request to %s, got %s", expectedURL, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	err := NotifyScorecard(svr.URL, DocID)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestClient_NotifyScorecardStatus(t *testing.T) {
	DocID := "SC:anonymous--submitted:20230322220711--2block:0:02/19/2023_20_00_-_03/21/2023_20_00"
	escapedDocID := url.PathEscape(DocID)
	expectedURL := fmt.Sprintf("/setStatusScorecard/%v", escapedDocID)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Path automatically unescapes URL-encoded paths so we need to use the RawPath here.
		if r.URL.RawPath != expectedURL {
			t.Errorf("Expected request to %s, got %s", expectedURL, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	err := NotifyScorecardStatus(svr.URL, DocID, "frog", nil)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	err = NotifyScorecardStatus(svr.URL, DocID, "frog", fmt.Errorf("error"))
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}
