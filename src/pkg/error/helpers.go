package er

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

const maxContextLen = 64 * 1024 // hard cap to keep logs/emails sane

// stringifyContext tries to produce a readable, compact string for common types,
// falling back to %#v when needed, and truncating very large outputs.
func StringifyContext(ctx any) string {
	switch v := ctx.(type) {
	case string:
		return capLen(v)

	case []byte:
		// Treat byte slices as text if valid UTF-8 (e.g., debug.Stack()).
		if utf8.Valid(v) {
			return capLen(string(v))
		}
		// Otherwise, show as hex (truncated).
		const preview = 4096
		if len(v) > preview {
			return fmt.Sprintf("bytes(%d): %x… (truncated)", len(v), v[:preview])
		}
		return fmt.Sprintf("%x", v)

	case json.RawMessage:
		if json.Valid(v) {
			return capLen(prettyOrCompactJSON(v))
		}
		return capLen(string(v))

	case error:
		return capLen(v.Error())

	case fmt.Stringer:
		return capLen(v.String())

	default:
		// Try JSON for maps/slices/structs to get stable-ish, readable output.
		if js, ok := tryJSON(v); ok {
			return capLen(js)
		}
		// Fallback to Go %#v dump.
		return capLen(fmt.Sprintf("%#v", v))
	}
}

func tryJSON(v any) (string, bool) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	// No indent first; if small, we’ll pretty it later.
	if err := enc.Encode(v); err != nil {
		return "", false
	}
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1] // json.Encoder adds a trailing newline
	}
	// If reasonably small, return a pretty version for readability.
	if len(b) <= 2048 {
		var out bytes.Buffer
		if err := json.Indent(&out, b, "", "  "); err == nil {
			return out.String(), true
		}
	}
	return string(b), true
}

func prettyOrCompactJSON(b []byte) string {
	if len(b) <= 2048 {
		var out bytes.Buffer
		if err := json.Indent(&out, b, "", "  "); err == nil {
			return out.String()
		}
	}
	return string(b)
}

func capLen(s string) string {
	if len(s) <= maxContextLen {
		return s
	}
	return s[:maxContextLen] + fmt.Sprintf("… (truncated to %d of %d bytes)", maxContextLen, len(s))
}
