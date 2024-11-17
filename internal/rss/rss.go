package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type Item struct {
	GUID        string   `xml:"guid"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
	Categogies  []string `xml:"category"`
}

func FetchFeed(url string) (string, error) {
	// Fetch the RSS feed from the url and return the content
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching RSS feed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading RSS feed: %v\n", err)
		return "", err
	}

	return string(body), nil
}

func ParseFeed(content string) (RSS, error) {
	// Parse the RSS feed content and return the RSS struct
	var rss RSS
	err := xml.Unmarshal([]byte(content), &rss)
	if err != nil {
		fmt.Printf("Error parsing RSS feed: %v\n", err)
		return RSS{}, err
	}

	return rss, nil
}

// func dumpChannel(rss RSS) {
// 	for _, item := range rss.Channel.Items {
// 		fmt.Printf("GUID: %s\n", item.GUID)
// 		fmt.Printf("Title: %s\n", item.Title)
// 		fmt.Printf("Description: %s\n", item.Description)
// 		fmt.Printf("Link: %s\n", item.Link)
// 		fmt.Printf("PubDate: %s\n", item.PubDate)
// 		fmt.Println()
// 	}
// }
