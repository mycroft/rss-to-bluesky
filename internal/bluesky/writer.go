package bluesky

import (
	"fmt"

	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

func (bs *BlueskyClient) WriteBlueskyPost(item rss.Item) (bool, error) {
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

func (bs *BlueskyClient) WriteBlueskyPosts(rss rss.RSS) error {
	number := 0

	err := bs.CheckSession()
	if err != nil {
		panic(err)
	}

	for _, item := range rss.Channel.Items {
		found, err := bs.DB.Has(item.GUID)
		if err != nil {
			return err
		}

		if found {
			continue
		}

		if bs.DryRun {
			number += 1

			fmt.Printf("Would write post: %s (ts: %s)\n", item.Title, item.PubDate)

			if number >= bs.Number {
				break
			}
			continue
		}

		number += 1

		written, err := bs.WriteBlueskyPost(item)
		if err != nil {
			return err
		}

		if written {
			err = bs.DB.Set(item.GUID, []byte("1"))
			if err != nil {
				return err
			}
		}

		if number >= bs.Number {
			break
		}
	}

	return nil
}
