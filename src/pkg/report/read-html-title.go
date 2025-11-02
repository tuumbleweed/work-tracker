package report

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
)

/*
ReadHTMLTitleFromBytes reads the first non-empty <title>...</title> from an
already-loaded HTML document contained in data.

Returns:
  - title string
  - e *er.Error (nil on success)
*/
func ReadHTMLTitleFromBytes(data []byte) (title string, e *er.Error) {
	logger.Log(logger.Notice, logger.BlueColor, "%s (%s %d)", "Reading HTML title from bytes", "len", len(data))

	var reader *bytes.Reader
	reader = bytes.NewReader(data)

	title, e = readHTMLTitle(reader)
	if e != nil {
		return
	}

	logger.Log(logger.Info1, logger.GreenColor, "%s '%s'", "Found HTML title", title)
	return
}

/*
readHTMLTitle scans an HTML stream and returns the first non-empty <title> text.
It does not assume <head> placement and tolerates whitespace or multi-chunk text.
*/
func readHTMLTitle(r io.Reader) (title string, e *er.Error) {
	logger.Log(logger.Debug3, logger.BlueColor, "%s", "Scanning HTML stream for <title>")

	var tokenizer *html.Tokenizer
	tokenizer = html.NewTokenizer(r)

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			parseErr := tokenizer.Err()
			if parseErr == io.EOF {
				e = er.NewError(nil, "no <title> element found", "")
				e.Print(er.ErrorTypeWarning, logger.Warning1, 0)
				return
			}
			e = er.NewError(parseErr, "tokenizer error while reading HTML", "")
			e.Print(er.ErrorTypeError, logger.Error1, 0)
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
								logger.Log(logger.Info2, logger.GreenColor, "%s '%s'", "Captured title", title)
								return
							}
							logger.Log(logger.Notice, logger.CyanColor, "%s", "Encountered empty <title>, continuing scan")
							break
						}
					}

					if innerType == html.ErrorToken {
						err := tokenizer.Err()
						if err == io.EOF {
							title = strings.TrimSpace(builder.String())
							if title != "" {
								logger.Log(logger.Info2, logger.GreenColor, "%s '%s'", "Captured title at EOF", title)
								return
							}
							e = er.NewError(nil, "no <title> element found before EOF", "")
							e.Print(er.ErrorTypeWarning, logger.Warning1, 0)
							return
						}
						e = er.NewError(err, "tokenizer error while reading <title> contents", "")
						e.Print(er.ErrorTypeError, logger.Error1, 0)
						return
					}
				}
			}
		}
	}
}
