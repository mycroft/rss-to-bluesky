package bluesky

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mycroft/rss-to-bluesky/internal/db"
)

type BlueskyClient struct {
	Ready          bool
	Session        Session
	DB             *db.DB
	DryRun         bool
	Number         int
	IgnoreExisting bool
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

func NewClient(db *db.DB, dryRun bool, number int, ignoreExisting bool) BlueskyClient {
	return BlueskyClient{
		Session:        Session{},
		Ready:          false,
		DB:             db,
		DryRun:         dryRun,
		Number:         number,
		IgnoreExisting: ignoreExisting,
	}
}

// Fetches image hosted at `source_url` and uploads it to bsky servers.
// Returns a reference to this upload to be embeded in a post.
func (bs *BlueskyClient) UploadBlob(source_url string) (Blob, error) {
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
	req.Header.Set("Authorization", "Bearer "+bs.Session.AccessJWT)

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
