package bluesky

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	// bsky rejects a post record whose embedded blob is over 1000000 bytes.
	// Stay under that so the post itself never fails on the thumbnail.
	thumbnailBudget = 900_000

	// Upper bound on what we are willing to pull down for a card thumbnail.
	maxDownloadSize = 20 << 20

	// Card thumbnails render small, so anything larger is wasted bytes.
	maxThumbnailDimension = 1200
)

// shrinkImage downscales and re-encodes an oversized preview image so it fits
// under thumbnailBudget. It returns the encoded image and its mime type.
func shrinkImage(data []byte) ([]byte, string, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decoding image: %v", err)
	}

	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, "", fmt.Errorf("image has empty bounds")
	}

	if width > maxThumbnailDimension || height > maxThumbnailDimension {
		if width > height {
			height = height * maxThumbnailDimension / width
			width = maxThumbnailDimension
		} else {
			width = width * maxThumbnailDimension / height
			height = maxThumbnailDimension
		}
	}

	// Extreme aspect ratios can round the short side down to zero.
	width = max(width, 1)
	height = max(height, 1)

	// JPEG carries no alpha channel, so flatten onto white first: without this
	// a transparent PNG re-encodes with a black background.
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	// Step the quality down until it fits. A 1200px JPEG is well under budget
	// at 85, so the lower rungs are only a safety net.
	for _, quality := range []int{85, 70, 55, 40} {
		buffer := &bytes.Buffer{}
		if err := jpeg.Encode(buffer, dst, &jpeg.Options{Quality: quality}); err != nil {
			return nil, "", fmt.Errorf("encoding jpeg: %v", err)
		}

		if buffer.Len() <= thumbnailBudget {
			return buffer.Bytes(), "image/jpeg", nil
		}
	}

	return nil, "", fmt.Errorf("still over %d bytes at lowest quality", thumbnailBudget)
}
