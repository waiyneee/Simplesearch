package crawler

import (
	"bytes"
	"net/url"

	"strings"

	"golang.org/x/net/html"
)

func isArticleLink(link string) bool {
	//to check whether or not it is article link
	//we hav to check if a link is valid article link
	//or just another noise
	// badPatterns := []string{
	// 	"(disambiguation)",
	// 	"/wiki/Main_Page",
	// 	"/wiki/Help:",
	// 	"/wiki/Wikipedia:",
	// 	"/wiki/File:",
	// 	"/wiki/Category:",
	// 	"/wiki/Special:",
	// 	"/wiki/Template:",
	// 	"/wiki/Portal:",
	// 	"/wiki/Talk:",
	// 	"/wiki/Draft:",
	// 	"/wiki/Module:",
	// 	"/wiki/Book:",
	// 	"/wiki/TimedText:",
	// 	"/wiki/MediaWiki:",
	// 	"/wiki/Topic:",
	// }

	// for patterns :=range badPatterns{
	// 	if strings.Contains(link,patterns) {
	// 		return false
	// 	}
	// }
	if link == "" {
		return false
	}

	// Reject obvious non-content or malformed URLs
	if strings.Contains(link, "#") || strings.Contains(link, "?") {
		return false
	}

	if !strings.HasPrefix(link, "https://en.wikipedia.org/wiki/") {
		return false
	}

	// Extract the title part after /wiki/
	title := strings.TrimPrefix(link, "https://en.wikipedia.org/wiki/")

	// Skip special pages like Main_Page
	if title == "Main_Page" {
		return false
	}

	// Reject namespace pages and talk pages.
	// Anything with ':' in the title is usually not a normal article.
	if strings.Contains(title, ":") {
		return false
	}

	// Still keep disambiguation pages out
	if strings.Contains(strings.ToLower(title), "(disambiguation)") {
		return false
	}

	return true

}

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

					if isArticleLink(href) {
						seen[href] = struct{}{}
						links = append(links, href)
					}
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
