package crawler

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// ExtractPageContent extracts <title> and visible text from HTML.
// It skips script/style/noscript text and collapses whitespace.
func ExtractPageContent(body []byte) (title string, text string, err error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}

	var parts []string
	parts = make([]string, 0, 256)

	var walk func(*html.Node, bool)
	walk = func(n *html.Node, skip bool) {
		if n == nil {
			return
		}

		// Skip non-content nodes
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript":
				skip = true
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			}
		}

		if !skip && n.Type == html.TextNode {
			s := strings.TrimSpace(n.Data)
			if s != "" {
				parts = append(parts, s)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, skip)
		}
	}

	walk(doc, false)

	text = strings.Join(parts, " ")
	text = strings.Join(strings.Fields(text), " ")
	title = strings.Join(strings.Fields(strings.TrimSpace(title)), " ")

	return title, text, nil
}
