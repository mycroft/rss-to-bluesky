package bluesky

import (
	"fmt"
	"time"
)

func ConvertPubDateToRFC3339(pubDate string) (string, error) {
	t, err := time.Parse(time.RFC1123Z, pubDate)
	if err != nil {
		fmt.Printf("Error parsing pubDate: %v\n", err)
		return "", err
	}

	// Format the time to RFC3339
	rfc3339 := t.Format(time.RFC3339)

	return rfc3339, nil
}
