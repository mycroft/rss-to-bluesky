package bluesky

import (
	"fmt"
	"log"

	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

func (bs *BlueskyClient) WriteBlueskyPost(item rss.Item) (bool, error) {
	// add a # for each items in item.Categories
	categories := make([]string, len(item.Categories))
	for i, category := range item.Categories {
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
			blob_ref, err := bs.UploadBlob(meta_info.ThumbnailUrl)
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
	err = bs.SendPost(item.GUID, content, facets, embed, timestamp)
	if err != nil {
		return false, err
	}

	return true, nil
}

// limitReached reports whether enough posts have been written for this run.
// A non-positive Number means no limit.
func (bs *BlueskyClient) limitReached(posted int) bool {
	return bs.Number > 0 && posted >= bs.Number
}

func (bs *BlueskyClient) WriteBlueskyPosts(rss rss.RSS) error {
	posted := 0
	failed := 0

	if err := bs.CheckSession(); err != nil {
		return fmt.Errorf("error checking session: %v", err)
	}

	// A single bad item must not strand the rest of the feed: successes are
	// recorded as they happen, per-item failures are logged and skipped, and
	// only the setup and database errors below abort the run.
	write := bs.writePost
	if write == nil {
		write = bs.WriteBlueskyPost
	}

	for _, item := range rss.Channel.Items {
		found, err := bs.DB.Has(item.GUID)
		if err != nil {
			return fmt.Errorf("error reading %s from database: %v", item.GUID, err)
		}

		if found {
			continue
		}

		if bs.DryRun {
			posted += 1

			fmt.Printf("Would write post: %s (ts: %s)\n", item.Title, item.PubDate)

			if bs.limitReached(posted) {
				break
			}
			continue
		}

		written, err := write(item)
		if err != nil {
			failed += 1
			log.Printf("skipping %s: %v", item.GUID, err)
			continue
		}

		if written {
			// A post that cannot be recorded would be posted again on the
			// next run, so stop rather than duplicate the rest of the feed.
			if err := bs.DB.Set(item.GUID, []byte("1")); err != nil {
				return fmt.Errorf("error recording %s in database: %v", item.GUID, err)
			}

			posted += 1
		}

		if bs.limitReached(posted) {
			break
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d item(s) failed, %d posted", failed, posted)
	}

	return nil
}
