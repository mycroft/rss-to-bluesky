package main

import (
	"flag"

	"github.com/mycroft/rss-to-bluesky/internal/bluesky"
	"github.com/mycroft/rss-to-bluesky/internal/db"
	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

var (
	dryRun         bool
	ignoreExisting bool
	number         int
)

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run mode, do not post to bluesky")
	flag.BoolVar(&ignoreExisting, "ignore-existing", false, "Ignore existing posts in database")
	flag.IntVar(&number, "number", -1, "Number of posts to check")
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

	db, err := db.Open()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	bs := bluesky.NewClient(db, dryRun, number, ignoreExisting)
	bs.CheckSession()

	// some code to test the bluesky client
	// err = bs.GetUser()
	// fmt.Println(err)

	err = bs.WriteBlueskyPosts(rss)
	if err != nil {
		panic(err)
	}
}
