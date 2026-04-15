package crawler

import (
	"bytes"
	"net/url"

	"golang.org/x/net/html"
)

func ExtractLinks(body []byte, base *url.URL) ([]string, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var links []string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				href := normalizeWikiLink(attr.Val, base)
				if href == "" {
					break
				}

				if _, ok := seen[href]; !ok {
					seen[href] = struct{}{}
					links = append(links, href)
				}
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