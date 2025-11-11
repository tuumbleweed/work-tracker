package report

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

/*
ReadHTMLTitleFromBytes reads the first non-empty <title>...</title> from an
already-loaded HTML document contained in data.

Returns:
  - title string
  - e *xerr.Error (nil on success)
*/
func ReadHTMLTitleFromBytes(data []byte) (title string, e *xerr.Error) {
	tl.Log(tl.Notice, palette.Blue, "%s (%s %d)", "Reading HTML title from bytes", "len", len(data))

	reader := bytes.NewReader(data)

	title, e = readHTMLTitle(reader)
	if e != nil {
		return
	}

	tl.Log(tl.Info1, palette.Green, "%s '%s'", "Found HTML title", title)
	return
}

/*
readHTMLTitle scans an HTML stream and returns the first non-empty <title> text.
It does not assume <head> placement and tolerates whitespace or multi-chunk text.
*/
func readHTMLTitle(r io.Reader) (title string, e *xerr.Error) {
	tl.Log(tl.Debug3, palette.Blue, "%s", "Scanning HTML stream for <title>")

	tokenizer := html.NewTokenizer(r)

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			parseErr := tokenizer.Err()
			if parseErr == io.EOF {
				e = xerr.NewError(nil, "no <title> element found", "")
				e.Print(xerr.ErrorTypeWarning, tl.Warning1, 0)
				return
			}
			e = xerr.NewError(parseErr, "tokenizer error while reading HTML", "")
			e.Print(xerr.ErrorTypeError, tl.Error1, 0)
			return
		}

		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			token := tokenizer.Token()
			if strings.EqualFold(token.Data, "title") {
				var builder strings.Builder

				for {
					innerType := tokenizer.Next()

					if innerType == html.TextToken {
						builder.Write(tokenizer.Text())
						continue
					}

					if innerType == html.EndTagToken {
						endTok := tokenizer.Token()
						if strings.EqualFold(endTok.Data, "title") {
							candidate := strings.TrimSpace(builder.String())
							if candidate != "" {
								title = candidate
								tl.Log(tl.Info2, palette.Green, "%s '%s'", "Captured title", title)
								return
							}
							tl.Log(tl.Notice, palette.Cyan, "%s", "Encountered empty <title>, continuing scan")
							break
						}
					}

					if innerType == html.ErrorToken {
						err := tokenizer.Err()
						if err == io.EOF {
							title = strings.TrimSpace(builder.String())
							if title != "" {
								tl.Log(tl.Info2, palette.Green, "%s '%s'", "Captured title at EOF", title)
								return
							}
							e = xerr.NewError(nil, "no <title> element found before EOF", "")
							e.Print(xerr.ErrorTypeWarning, tl.Warning1, 0)
							return
						}
						e = xerr.NewError(err, "tokenizer error while reading <title> contents", "")
						e.Print(xerr.ErrorTypeError, tl.Error1, 0)
						return
					}
				}
			}
		}
	}
}
