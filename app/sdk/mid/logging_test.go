package mid

import "testing"

func TestScrubQuery(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no sensitive params", "page=1&rows=50", "page=1&rows=50"},
		{"token at start", "token=aaa.bbb.ccc", "token=REDACTED"},
		{"token among others", "page=1&token=aaa.bbb&rows=50", "page=1&token=REDACTED&rows=50"},
		{"access_token redacted", "access_token=secret", "access_token=REDACTED"},
		{"case-insensitive key", "Token=secret", "Token=REDACTED"},
		{"key substring is not redacted", "mytoken=keepme", "mytoken=keepme"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scrubQuery(tt.in); got != tt.want {
				t.Errorf("scrubQuery(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
