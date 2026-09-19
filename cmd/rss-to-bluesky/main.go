package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mycroft/rss-to-bluesky/internal/bluesky"
	"github.com/mycroft/rss-to-bluesky/internal/db"
	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

var (
	dryRun      bool
	number      int
	showVersion bool
)

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run mode, do not post to bluesky")
	flag.IntVar(&number, "number", -1, "Maximum number of posts to write (-1 for no limit)")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Println(buildVersion())
		return
	}

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
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	bs := bluesky.NewClient(db, dryRun, number)
	if err := bs.CheckSession(); err != nil {
		panic(err)
	}

	// some code to test the bluesky client
	// err = bs.GetUser()
	// fmt.Println(err)

	err = bs.WriteBlueskyPosts(rss)
	if err != nil {
		panic(err)
	}
}
