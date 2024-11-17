package bluesky

import (
	"fmt"
	"os"

	"github.com/mycroft/rss-to-bluesky/internal/db"
	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

func WriteBlueskyPost(session Session, item rss.Item) (bool, error) {
	// add a # for each items in item.Categogies
	categories := make([]string, len(item.Categogies))
	for i, category := range item.Categogies {
		categories[i] = "#" + category
	}

	content := fmt.Sprintf("%s ", item.Title)
	guidFacet := Facet{
		Index: FacetIndex{
			ByteStart: len(content),
			ByteEnd:   len(content) + len(item.GUID),
		},
		Features: []FacetFeature{
			{
				Type: "app.bsky.richtext.facet#link",
				Uri:  item.GUID,
			},
		},
	}

	content += fmt.Sprintf("%s ", item.GUID)

	facets := []Facet{
		guidFacet,
	}

	for _, category := range categories {
		categoryFacet := Facet{
			Index: FacetIndex{
				ByteStart: len(content),
				ByteEnd:   len(content) + len(category),
			},
			Features: []FacetFeature{
				{
					Type: "app.bsky.richtext.facet#tag",
					Tag:  category,
				},
			},
		}

		content += fmt.Sprintf("%s ", category)
		facets = append(facets, categoryFacet)
	}

	// Add the item link
	linkFacet := Facet{
		Index: FacetIndex{
			ByteStart: len(content),
			ByteEnd:   len(content) + len(item.Link),
		},
		Features: []FacetFeature{
			{
				Type: "app.bsky.richtext.facet#link",
				Uri:  item.Link,
			},
		},
	}

	content += item.Link
	facets = append(facets, linkFacet)

	timestamp, err := ConvertPubDateToRFC3339(item.PubDate)
	if err != nil {
		return false, err
	}

	// Write a post to bluesky.social
	err = SendPost(session, content, facets, timestamp)
	if err != nil {
		return false, err
	}

	return true, nil
}

func WriteBlueskyPosts(rss rss.RSS) error {
	session := CheckBlueskySession()

	dbInstance, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer dbInstance.Close()

	for _, item := range rss.Channel.Items {
		found, err := db.Has(dbInstance, item.GUID)
		if err != nil {
			return err
		}

		if found {
			continue
		}

		if session.AccessJWT == "" {
			session, err = GetBlueskySession(os.Getenv("BLUESKY_USER"), os.Getenv("BLUESKY_PASS"))
			if err != nil {
				return err
			}
		}

		written, err := WriteBlueskyPost(session, item)
		if err != nil {
			return err
		}

		if written {
			err = db.Set(dbInstance, item.GUID, "1")
			if err != nil {
				return err
			}
		}
	}

	// Write all the posts from the RSS feed to bluesky.social
	return nil
}
