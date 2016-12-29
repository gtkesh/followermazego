package main

import "testing"

func TestParseEvent(t *testing.T) {
	tests := []struct {
		eventStr string
		e        event
		err      bool
	}{
		// Good inputs.
		{"666|F|60|50", event{666, Follow, 60, 50, "666|F|60|50"}, false},
		{"1|U|12|9", event{1, Unfollow, 12, 9, "1|U|12|9"}, false},
		{"542532|B", event{542532, Broadcast, 0, 0, "542532|B"}, false},
		{"43|P|32|56", event{43, PrivateMessage, 32, 56, "43|P|32|56"}, false},
		{"634|S|32", event{634, StatusUpdate, 32, 0, "634|S|32"}, false},

		// Bad inputs.
		{"", event{}, true},
	}

	for _, tt := range tests {
		e, err := parseEvent(tt.eventStr)
		if err != nil {
			if !tt.err {
				t.Errorf("expected parse error, got no error\n", err)
			}
		}
		if e != tt.e {
			t.Error(
				"expected: ", tt.e,
				"got: ", e,
			)
		}
	}
}
