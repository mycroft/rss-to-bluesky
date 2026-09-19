package bluesky

import (
	"errors"
	"strings"
	"testing"

	"github.com/mycroft/rss-to-bluesky/internal/rss"
)

func TestLimitReached(t *testing.T) {
	tests := []struct {
		name   string
		number int
		posted int
		want   bool
	}{
		{"default is unlimited", -1, 1, false},
		{"default stays unlimited later in the run", -1, 500, false},
		{"zero is unlimited", 0, 3, false},
		{"under the limit", 5, 4, false},
		{"at the limit", 5, 5, true},
		{"over the limit", 5, 6, true},
		{"limit of one", 1, 1, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := BlueskyClient{Number: test.number}
			if got := bs.limitReached(test.posted); got != test.want {
				t.Errorf("Number=%d posted=%d: got %v, want %v",
					test.number, test.posted, got, test.want)
			}
		})
	}
}

// fakeStore records what the loop writes, and can fail on demand.
type fakeStore struct {
	seen   map[string]bool
	setErr error
}

func newFakeStore(seen ...string) *fakeStore {
	s := &fakeStore{seen: map[string]bool{}}
	for _, guid := range seen {
		s.seen[guid] = true
	}
	return s
}

func (s *fakeStore) Get(key string) ([]byte, error) { return nil, nil }
func (s *fakeStore) Has(key string) (bool, error)   { return s.seen[key], nil }

func (s *fakeStore) Set(key string, value []byte) error {
	if s.setErr != nil {
		return s.setErr
	}
	s.seen[key] = true
	return nil
}

func feed(guids ...string) rss.RSS {
	items := make([]rss.Item, len(guids))
	for i, guid := range guids {
		items[i] = rss.Item{GUID: guid, Title: guid}
	}
	return rss.RSS{Channel: rss.Channel{Items: items}}
}

// The point of the change: a failing item is skipped, not fatal, and the
// items after it still go out.
func TestFailingItemDoesNotStopTheRun(t *testing.T) {
	store := newFakeStore()
	attempted := []string{}

	bs := BlueskyClient{Ready: true, DB: store}
	bs.writePost = func(item rss.Item) (bool, error) {
		attempted = append(attempted, item.GUID)
		if item.GUID == "b" {
			return false, errors.New("unparseable pubDate")
		}
		return true, nil
	}

	err := bs.WriteBlueskyPosts(feed("a", "b", "c"))
	if err == nil {
		t.Fatal("expected the run to report the failed item")
	}

	if got, want := strings.Join(attempted, ","), "a,b,c"; got != want {
		t.Errorf("attempted %q, want %q", got, want)
	}

	if store.seen["b"] {
		t.Error("failed item was recorded as posted")
	}
	for _, guid := range []string{"a", "c"} {
		if !store.seen[guid] {
			t.Errorf("%q was not recorded as posted", guid)
		}
	}
}

// Failures must not eat into -number: the limit counts posts written.
func TestLimitCountsOnlySuccesses(t *testing.T) {
	store := newFakeStore()
	posted := []string{}

	bs := BlueskyClient{Ready: true, DB: store, Number: 2}
	bs.writePost = func(item rss.Item) (bool, error) {
		if item.GUID == "a" {
			return false, errors.New("nope")
		}
		posted = append(posted, item.GUID)
		return true, nil
	}

	if err := bs.WriteBlueskyPosts(feed("a", "b", "c", "d")); err == nil {
		t.Fatal("expected the run to report the failed item")
	}

	if got, want := strings.Join(posted, ","), "b,c"; got != want {
		t.Errorf("posted %q, want %q", got, want)
	}
}

// Already-posted items are skipped without being attempted again.
func TestKnownItemsAreSkipped(t *testing.T) {
	attempted := 0

	bs := BlueskyClient{Ready: true, DB: newFakeStore("a", "b")}
	bs.writePost = func(item rss.Item) (bool, error) {
		attempted += 1
		return true, nil
	}

	if err := bs.WriteBlueskyPosts(feed("a", "b")); err != nil {
		t.Fatalf("WriteBlueskyPosts: %v", err)
	}

	if attempted != 0 {
		t.Errorf("attempted %d known items, want 0", attempted)
	}
}

// A post that cannot be recorded would be reposted next run, so it stays fatal.
func TestUnrecordablePostStopsTheRun(t *testing.T) {
	store := newFakeStore()
	store.setErr = errors.New("disk full")
	attempted := 0

	bs := BlueskyClient{Ready: true, DB: store}
	bs.writePost = func(item rss.Item) (bool, error) {
		attempted += 1
		return true, nil
	}

	if err := bs.WriteBlueskyPosts(feed("a", "b", "c")); err == nil {
		t.Fatal("expected a database error to abort the run")
	}

	if attempted != 1 {
		t.Errorf("attempted %d items, want 1", attempted)
	}
}
