package zpl

import "strings"

// Sanitize strips ZPL command-prefix characters from user-supplied
// strings before they are interpolated into ^FD data fields.
//
// ZPL's default format-command prefix is `^` (e.g. ^XA / ^FD / ^FS)
// and default control-command prefix is `~` (e.g. ~JS / ~HI). Inside
// a ^FD field an attacker-controlled `^FS` would terminate the field
// and allow arbitrary subsequent commands (^XZ^XA...). Stripping
// both prefixes neutralizes that path. The `,` parameter delimiter
// is intentionally preserved: it has no command effect inside ^FD
// and appears legitimately in names like "Acme, Inc.".
func Sanitize(s string) string {
	if !strings.ContainsAny(s, "^~") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '^' || r == '~' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// SanitizePtr returns nil when p is nil, otherwise a pointer to the
// sanitized copy of *p.
func SanitizePtr(p *string) *string {
	if p == nil {
		return nil
	}
	s := Sanitize(*p)
	return &s
}
