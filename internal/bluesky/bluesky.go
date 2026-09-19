package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mycroft/rss-to-bluesky/internal/httpx"
	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

// Store is the part of the database this package needs.
type Store interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Has(key string) (bool, error)
}

type BlueskyClient struct {
	Ready   bool
	Session Session
	DB      Store
	DryRun  bool
	Number  int

	// writePost is the per-item write. It is nil outside tests, where
	// WriteBlueskyPost is used instead.
	writePost func(item rss.Item) (bool, error)
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

type UploadBlobResponse struct {
	Blob Blob `json:"blob"`
}

func NewClient(db Store, dryRun bool, number int) BlueskyClient {
	return BlueskyClient{
		Session: Session{},
		Ready:   false,
		DB:      db,
		DryRun:  dryRun,
		Number:  number,
	}
}

// Fetches image hosted at `source_url` and uploads it to bsky servers.
// Returns a reference to this upload to be embeded in a post.
func (bs *BlueskyClient) UploadBlob(source_url string) (Blob, error) {
	image_resp, err := httpx.Get(source_url)
	if err != nil {
		fmt.Printf("Error loading preview image: %v\n", err)
		return Blob{}, err
	}
	defer image_resp.Body.Close()

	if image_resp.StatusCode != http.StatusOK {
		return Blob{}, fmt.Errorf("preview image request returned %s", image_resp.Status)
	}

	if image_resp.ContentLength > maxDownloadSize {
		return Blob{}, fmt.Errorf("preview image is %d bytes, refusing to download", image_resp.ContentLength)
	}

	image_data, err := io.ReadAll(io.LimitReader(image_resp.Body, maxDownloadSize+1))
	if err != nil {
		fmt.Printf("Error reading preview image: %v\n", err)
		return Blob{}, err
	}

	if len(image_data) > maxDownloadSize {
		return Blob{}, fmt.Errorf("preview image exceeds %d bytes", maxDownloadSize)
	}

	mime_type, err := resolveImageMimeType(image_data, image_resp.Header.Get("Content-Type"))
	if err != nil {
		return Blob{}, fmt.Errorf("preview image at %s: %v", source_url, err)
	}

	// The blob size limit is only enforced when the post record is created, so
	// an oversized upload fails the whole post and orphans the blob. Shrink it
	// here instead, and let the caller drop the thumbnail if that is not possible.
	if len(image_data) > thumbnailBudget {
		resized, resized_mime, err := shrinkImage(image_data)
		if err != nil {
			return Blob{}, fmt.Errorf("preview image is %d bytes and could not be shrunk: %v", len(image_data), err)
		}

		image_data = resized
		mime_type = resized_mime
	}

	url := "https://bsky.social/xrpc/com.atproto.repo.uploadBlob"
	req, err := http.NewRequest("POST", url, bytes.NewReader(image_data))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		return Blob{}, err
	}

	req.Header.Set("Content-Type", mime_type)
	req.Header.Set("Authorization", "Bearer "+bs.Session.AccessJWT)

	upload_resp, err := httpx.Client.Do(req)
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

	if upload_resp.StatusCode != http.StatusOK {
		return Blob{}, fmt.Errorf("uploadBlob returned %s: %s", upload_resp.Status, string(body))
	}

	parsed_response := UploadBlobResponse{}
	if err := json.Unmarshal(body, &parsed_response); err != nil {
		fmt.Printf("Error unmarshaling upload response: %v\n", err)
		return Blob{}, err
	}

	return parsed_response.Blob, nil

}

func (bs *BlueskyClient) SendPost(guid, content string, facets []Facet, embed Embed, pubDate string) error {
	post := Post{
		Type:      "app.bsky.feed.post",
		Text:      content,
		CreatedAt: pubDate,
		Langs:     []string{"en"},
		Facets:    facets,
		Embed:     embed,
	}

	postRequest := PostRequest{
		Repo:       bs.Session.DID,
		Collection: "app.bsky.feed.post",
		Record:     post,
	}

	endpoint := "/xrpc/com.atproto.repo.createRecord"

	body, err := bs.query("POST", endpoint, postRequest)
	if err != nil {
		return fmt.Errorf("error querying bluesky.social: %v", err)
	}

	log.Printf("GUID:%s (%s) res:%s\n", guid, pubDate, string(body))

	return nil
}
