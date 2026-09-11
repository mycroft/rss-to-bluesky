package bluesky

import "testing"

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
