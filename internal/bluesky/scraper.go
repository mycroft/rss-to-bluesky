package bluesky

import (
	"fmt"
	"net/http"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type MetaInfo struct {
	Title        string
	Description  string
	ThumbnailUrl string
}

func FindMetaProperty(root_node html.Node, property_name string) (string, error) {
	for node := range root_node.Descendants() {
		if node.Type == html.ElementNode && node.DataAtom == atom.Meta {
			found := false
			property := ""
			for _, attr := range node.Attr {
				if attr.Key == "property" && attr.Val == property_name {
					found = true
				}
				if attr.Key == "content" {
					property = attr.Val
				}
			}

			if found && property != "" {
				return property, nil
			}
		}
	}
	return "", fmt.Errorf("couldn't find property '%s'", property_name)
}

func FetchLinkMetaInfo(link string) (MetaInfo, error) {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("Error fetching from link: %v\n", err)
		return MetaInfo{}, err
	}
	defer resp.Body.Close()

	root_node, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("Couldn't parse html: %v\n", err)
		return MetaInfo{}, err
	}

	image_url, _ := FindMetaProperty(*root_node, "og:image")
	title, _ := FindMetaProperty(*root_node, "og:title")
	description, _ := FindMetaProperty(*root_node, "og:description")

	return MetaInfo{
		Title:        title,
		Description:  description,
		ThumbnailUrl: image_url,
	}, nil
}
