package client

import (
	"bytes"
	"fmt"
	"net/http"
)

// UpdateMATS creates an empty POST request to <appURL>/refreshScorecard/<docID>
// in order to inform MATS that a scorecard document has been updated
func UpdateMATS(appURL string, docID string) error {
	postURL := fmt.Sprintf("%v/refreshScorecard/%v", appURL, docID)
	req, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(([]byte{})))
	if err != nil {
		return fmt.Errorf("client: could not create request: %s\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: error making http request: %s\n", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("client: got response code %d from %v", res.StatusCode, postURL)
	}

	return nil
}
