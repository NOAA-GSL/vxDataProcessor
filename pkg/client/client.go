package client

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

// NotifyScorecard creates an empty POST request to <baseURL>/refreshScorecard/<docID>
// in order to inform the MATS scorecard app that a scorecard document has been updated.
func NotifyScorecard(baseURL string, docID string) error {
	url := fmt.Sprintf("%v/refreshScorecard/%v", baseURL, url.PathEscape(docID))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(([]byte{})))
	if err != nil {
		return fmt.Errorf("client: could not create request: %s\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %s\n", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("client: got response code %d from %v", res.StatusCode, url)
	}

	return nil
}
