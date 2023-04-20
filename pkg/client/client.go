package client

import (
	"bytes"
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
		return fmt.Errorf("client: could not create request: %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %s", err)
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
func NotifyScorecardStatus(baseURL, docID string, err error) error {
	var doc []byte
	if err != nil {
		doc = []byte("{status:\"error\", error:" + err.Error() + "}")
	} else {
		doc = []byte("{status:\"ready\", error:\"none\"}")
	}
	scorecardURL := fmt.Sprintf("%v/setStatusScorecard/%v", baseURL, url.PathEscape(docID))
	req, err := http.NewRequest(http.MethodPost, scorecardURL, bytes.NewBuffer(doc))
	if err != nil {
		return fmt.Errorf("client: could not create request: %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("client: got response code %d from %v", res.StatusCode, scorecardURL)
	}

	return nil
}
