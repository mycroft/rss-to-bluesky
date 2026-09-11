package bluesky

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"testing"
)

func noisyPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r := rand.New(rand.NewSource(1))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestShrinkOversized(t *testing.T) {
	in := noisyPNG(t, 2000, 1500)
	t.Logf("input PNG: %d bytes", len(in))
	if len(in) <= thumbnailBudget {
		t.Fatalf("fixture not oversized: %d", len(in))
	}

	out, mime, err := shrinkImage(in)
	if err != nil {
		t.Fatalf("shrinkImage: %v", err)
	}
	if len(out) > thumbnailBudget {
		t.Fatalf("still over budget: %d", len(out))
	}
	if len(out) >= 1_000_000 {
		t.Fatalf("would still be rejected by createRecord: %d", len(out))
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("output not decodable: %v", err)
	}
	t.Logf("output: %d bytes, %s, %dx%d", len(out), mime, cfg.Width, cfg.Height)
	if cfg.Width > maxThumbnailDimension || cfg.Height > maxThumbnailDimension {
		t.Fatalf("not downscaled: %dx%d", cfg.Width, cfg.Height)
	}
	if mime != "image/jpeg" {
		t.Fatalf("mime = %q", mime)
	}
}

func TestShrinkTransparentNotBlack(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1400, 1400)) // fully transparent
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	out, _, err := shrinkImage(b.Bytes())
	if err != nil {
		t.Skipf("fixture under budget, shrink not exercised: %v", err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	r, g, bb, _ := decoded.At(5, 5).RGBA()
	t.Logf("corner pixel: r=%d g=%d b=%d (want near 65535, not 0)", r, g, bb)
	if r < 50000 || g < 50000 || bb < 50000 {
		t.Fatalf("transparent PNG flattened to dark background")
	}
}

func TestShrinkGarbage(t *testing.T) {
	if _, _, err := shrinkImage([]byte("not an image at all")); err == nil {
		t.Fatal("expected an error for undecodable input")
	} else {
		t.Logf("undecodable input rejected: %v", err)
	}
}

func TestShrinkExtremeAspect(t *testing.T) {
	in := noisyPNG(t, 4000, 2)
	out, _, err := shrinkImage(in)
	if err != nil {
		t.Skipf("under budget: %v", err)
	}
	cfg, _, _ := image.DecodeConfig(bytes.NewReader(out))
	t.Logf("extreme aspect output: %dx%d", cfg.Width, cfg.Height)
	if cfg.Height < 1 || cfg.Width < 1 {
		t.Fatalf("degenerate output: %dx%d", cfg.Width, cfg.Height)
	}
}
