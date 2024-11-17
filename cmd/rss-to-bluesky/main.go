package main

import (
	"github.com/mycroft/rss-to-bluesky/internal/bluesky"
	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

func main() {
	// Write a program that fetches a RSS feed and post the most recent item to bluesky.
	// The RSS feed is: http://feeds.bbci.co.uk/news/rss.xml
	feedUrl := "https://lobste.rs/newest.rss"
	content, err := rss.FetchFeed(feedUrl)
	if err != nil {
		panic(err)
	}

	rss, err := rss.ParseFeed(content)
	if err != nil {
		panic(err)
	}

	err = bluesky.WriteBlueskyPosts(rss)
	if err != nil {
		panic(err)
	}
}
