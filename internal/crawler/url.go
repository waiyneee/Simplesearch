package crawler

import (
	"net/url"
	"strings"
)

// normalizeWikiLink resolves href against base and returns an absolute URL.
// Returns empty string when URL is not a valid en.wikipedia article candidate.
func normalizeWikiLink(href string, base *url.URL) string {
	if href == "" || base == nil {
		return ""
	}

	u, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(u)

	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	if resolved.Host != "en.wikipedia.org" {
		return ""
	}
	if !strings.HasPrefix(resolved.Path, "/wiki/") {
		return ""
	}

	// Canonicalization for dedup
	resolved.Fragment = ""
	resolved.RawQuery = ""

	return resolved.String()
}

func isArticleLink(link string) bool {
	if link == "" {
		return false
	}

	if !strings.HasPrefix(link, "https://en.wikipedia.org/wiki/") {
		return false
	}

	// Extra safety
	if strings.Contains(link, "#") || strings.Contains(link, "?") {
		return false
	}

	title := strings.TrimPrefix(link, "https://en.wikipedia.org/wiki/")
	if title == "" || title == "Main_Page" {
		return false
	}

	// Namespace pages usually contain ':'
	if strings.Contains(title, ":") {
		return false
	}

	if strings.Contains(strings.ToLower(title), "(disambiguation)") {
		return false
	}

	return true
}