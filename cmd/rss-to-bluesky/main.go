package main

import (
	"flag"

	"github.com/mycroft/rss-to-bluesky/internal/bluesky"
	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

var (
	dryRun bool
	all    bool
	one    bool
)

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run mode, do not post to bluesky")
	flag.BoolVar(&all, "all", false, "Post all items from the feed, ignoring the current database state")
	flag.BoolVar(&one, "one", false, "Post only one item from the feed")

}

func main() {
	flag.Parse()

	feedUrl := "https://lobste.rs/newest.rss"
	content, err := rss.FetchFeed(feedUrl)
	if err != nil {
		panic(err)
	}

	rss, err := rss.ParseFeed(content)
	if err != nil {
		panic(err)
	}

	err = bluesky.WriteBlueskyPosts(rss, all, one, dryRun)
	if err != nil {
		panic(err)
	}
}
