package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// NotifyScorecard creates an empty POST request to <baseURL>/refreshScorecard/<docID>
// in order to inform the MATS scorecard app that a scorecard document has been updated.
func NotifyScorecard(baseURL, docID string) error {
	scorecardURL := fmt.Sprintf("%v/refreshScorecard/%v", baseURL, url.PathEscape(docID))
	req, err := http.NewRequest(http.MethodPost, scorecardURL, bytes.NewBuffer(([]byte{})))
	if err != nil {
		return fmt.Errorf("client: could not create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("client: got response code %d from %v", res.StatusCode, scorecardURL)
	}

	return nil
}

// NotifyScorecardStatus creates a POST request to <baseURL>/setStatusScorecard/<docID>
// in order to inform the MATS scorecard app that a scorecard document is finished being processed
// and what the resulting status is.
func NotifyScorecardStatus(baseURL, docID, status string, err error) error {
	type statDoc struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	if err == nil {
		err = fmt.Errorf("none")
	}
	errorStat := statDoc{Status: "status", Error: err.Error()}
	// marshall this into a json document
	jsonData, err := json.Marshal(errorStat)
	if err != nil {
		return fmt.Errorf("client: could not create status doc: %w", err)
	}
	scorecardURL := fmt.Sprintf("%v/setStatusScorecard/%v", baseURL, url.PathEscape(docID))
	req, err := http.NewRequest(http.MethodPost, scorecardURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return fmt.Errorf("client: could not create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("client: got response code %d from %v", res.StatusCode, scorecardURL)
	}

	return nil
}
