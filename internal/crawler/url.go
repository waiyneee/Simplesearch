package crawler

import (
	"net/url"
	"strings"
)

func normalizeWikiLink(href string, base *url.URL) string {
	if href == "" {
		return ""
	}

	if strings.HasPrefix(href, "#") {
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

	// Skip special pages like /wiki/Help:Contents
	if strings.Contains(resolved.Path, ":") {
		return ""
	}

	resolved.Fragment = ""
	return resolved.String()
}