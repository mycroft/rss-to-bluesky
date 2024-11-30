package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (bs *BlueskyClient) query(method, endpoint string, params interface{}) ([]byte, error) {
	var resp *http.Response
	var err error
	var postBody io.Reader
	var finalUrl string

	client := &http.Client{}

	baseUrl := "https://bsky.social"

	if method == http.MethodGet {
		queryParams := url.Values{}
		for key, value := range params.(map[string]string) {
			queryParams.Add(key, value)
		}

		finalUrl = fmt.Sprintf("%s%s?%s", baseUrl, endpoint, queryParams.Encode())
	} else if method == http.MethodPost {
		finalUrl = fmt.Sprintf("%s%s", baseUrl, endpoint)

		requestBody, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request payload: %v", err)
		}

		if params != nil {
			postBody = bytes.NewBuffer(requestBody)
		} else {
			postBody = nil
		}
	} else {
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	req, err := http.NewRequest(method, finalUrl, postBody)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bs.Session.AccessJWT)

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Println(string(body))
		return nil, fmt.Errorf("unexpected return code returned: %d", resp.StatusCode)
	}

	return body, nil
}
