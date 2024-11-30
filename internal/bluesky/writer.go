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

	timestamp, err := ConvertPubDateToRFC3339(item.PubDate)
	if err != nil {
		return false, err
	}

	external_embed := ExternalEmbed{
		Uri:   item.Link,
		Title: item.Title,
	}

	meta_info, err := FetchLinkMetaInfo(item.Link)
	if err == nil {
		if meta_info.ThumbnailUrl != "" {
			blob_ref, err := UploadBlob(session, meta_info.ThumbnailUrl)
			if err != nil {
				fmt.Printf("Couldn't upload image blob to bsky: %v\n", err)
			} else {
				external_embed.Thumb = &blob_ref
			}
		}
		// a page might not have enough meta information to create the card,
		// so we only overwrite the fallback if it does
		if meta_info.Title != "" {
			external_embed.Title = meta_info.Title
		}
		external_embed.Description = meta_info.Description
	}

	embed := Embed{
		Type:     "app.bsky.embed.external",
		External: external_embed,
	}

	// Write a post to bluesky.social
	err = SendPost(session, content, facets, embed, timestamp)
	if err != nil {
		return false, err
	}

	return true, nil
}

func WriteBlueskyPosts(rss rss.RSS, all, one, dryRun bool) error {
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

		if found && !all {
			continue
		}

		if session.AccessJWT == "" {
			session, err = GetBlueskySession(os.Getenv("BLUESKY_USER"), os.Getenv("BLUESKY_PASS"))
			if err != nil {
				return err
			}
		}

		if dryRun {
			fmt.Printf("Would write post: %s (ts: %s)\n", item.Title, item.PubDate)

			if one {
				break
			}
			continue
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

		if one {
			break
		}
	}

	// Write all the posts from the RSS feed to bluesky.social
	return nil
}
