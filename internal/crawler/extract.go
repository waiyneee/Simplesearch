package crawler

import (
	"bytes"
	"net/url"

	"golang.org/x/net/html"
)

// ExtractLinks parses HTML and returns unique, normalized article links.
func ExtractLinks(body []byte, base *url.URL) ([]string, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	links := make([]string, 0, 64)

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				link := normalizeWikiLink(attr.Val, base)
				if link == "" {
					break
				}
				if !isArticleLink(link) {
					break
				}
				if _, ok := seen[link]; ok {
					break
				}

				seen[link] = struct{}{}
				links = append(links, link)
				break
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return links, nil
}