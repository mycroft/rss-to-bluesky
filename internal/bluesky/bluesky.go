package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Session struct {
	DID        string `json:"did"`
	AccessJWT  string `json:"accessJwt"`
	RefreshJWT string `json:"refreshJwt"`
	Active     bool   `json:"active"`
	Handle     string `json:"handle"`
	Email      string `json:"email"`
}

type PostRequest struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	Record     Post   `json:"record"`
}

type Post struct {
	Type      string   `json:"$type"`
	Text      string   `json:"text"`
	CreatedAt string   `json:"createdAt"`
	Langs     []string `json:"langs"`
	Facets    []Facet  `json:"facets"`
	Embed     Embed    `json:"embed"`
}

type FacetIndex struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type FacetFeature struct {
	Type string `json:"$type"`
	Uri  string `json:"uri"`
	Tag  string `json:"tag"`
}

type Facet struct {
	Index    FacetIndex     `json:"index"`
	Features []FacetFeature `json:"features"`
}

// https://docs.bsky.app/docs/advanced-guides/posts#website-card-embeds

type Ref struct {
	Link string `json:"$link"`
}

type Blob struct {
	Type     string `json:"$type"`
	Ref      Ref    `json:"ref"`
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
}

type ExternalEmbed struct {
	Uri         string `json:"uri"`
	Thumb       *Blob  `json:"thumb,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Embed struct {
	Type     string        `json:"$type"`
	External ExternalEmbed `json:"external"`
}

type UploadBlobReponse struct {
	Blob Blob `json:"blob"`
}

func CheckBlueskySession() Session {
	token := os.Getenv("BLUESKY_ACCESSJWT")
	did := os.Getenv("BLUESKY_DID")

	session := Session{
		AccessJWT: token,
		DID:       did,
	}

	return session
}

func GetBlueskySession(handle, password string) (Session, error) {
	url := "https://bsky.social/xrpc/com.atproto.server.createSession"

	// Create a new session on bluesky.social
	payload := map[string]string{
		"identifier": handle,
		"password":   password,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling request payload: %v\n", err)
		return Session{}, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error making HTTP POST request: %v\n", err)
		return Session{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return Session{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received non-200 status code: %d\n", resp.StatusCode)
		fmt.Printf("Response body: %s\n", body)
		return Session{}, fmt.Errorf("invalid return code returned: %d", resp.StatusCode)
	}

	var responsePayload Session
	err = json.Unmarshal(body, &responsePayload)
	if err != nil {
		fmt.Printf("Error unmarshaling response body: %v\n", err)
		return Session{}, fmt.Errorf("error unmarshaling response body: %v", err)
	}

	return responsePayload, nil
}

// Fetches image hosted at `source_url` and uploads it to bsky servers.
// Returns a reference to this upload to be embeded in a post.
func UploadBlob(session Session, source_url string) (Blob, error) {

	image_resp, err := http.Get(source_url)
	if err != nil {
		fmt.Printf("Error loading preview image: %x\n", err)
		return Blob{}, err
	}

	mime_type := image_resp.Header.Get("Content-Type")

	url := "https://bsky.social/xrpc/com.atproto.repo.uploadBlob"
	req, err := http.NewRequest("POST", url, image_resp.Body)
	if err != nil {
		fmt.Printf("Error creating HTTP request: %x\n", err)
		return Blob{}, err
	}

	req.Header.Set("Content-Type", mime_type)
	req.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	client := &http.Client{}
	upload_resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v\n", err)
		return Blob{}, err
	}
	defer upload_resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(upload_resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return Blob{}, err
	}
	defer upload_resp.Body.Close()

	parsed_response := UploadBlobReponse{}
	json.Unmarshal(body, &parsed_response)

	return parsed_response.Blob, nil

}

func SendPost(session Session, content string, facets []Facet, embed Embed, pubDate string) error {
	post := Post{
		Type:      "app.bsky.feed.post",
		Text:      content,
		CreatedAt: pubDate,
		Langs:     []string{"en"},
		Facets:    facets,
		Embed:     embed,
	}

	postRequest := PostRequest{
		Repo:       session.DID,
		Collection: "app.bsky.feed.post",
		Record:     post,
	}

	requestBody, err := json.Marshal(postRequest)
	if err != nil {
		fmt.Printf("Error marshaling request payload: %v\n", err)
		return err
	}

	url := "https://bsky.social/xrpc/com.atproto.repo.createRecord"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Println(string(body))

	return nil
}
