package crawler

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractPageContent extracts <title> and cleaned article text from Wikipedia HTML.
// It prefers the main article container and removes non-content blocks.
func ExtractPageContent(body []byte) (title string, text string, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}

	// Title
	title = strings.TrimSpace(doc.Find("title").First().Text())
	title = strings.Join(strings.Fields(title), " ")

	// 1) Select article root (fallback chain)
	var root *goquery.Selection
	for _, sel := range []string{
		"#mw-content-text .mw-parser-output",
		"#mw-content-text",
		"main",
		"body",
	} {
		s := doc.Find(sel).First()
		if s.Length() > 0 {
			root = s
			break
		}
	}
	if root == nil || root.Length() == 0 {
		return title, "", nil
	}

	// 2) Remove noisy blocks (DOM-level)
	root.Find(strings.Join([]string{
		"style", "script", "noscript",
		"table.infobox", "table.navbox", "table.vertical-navbox", "table.metadata",
		".reflist", ".references", ".mw-references-wrap", "sup.reference",
		"#toc", ".toc",
		".mw-editsection", ".hatnote",
		".navbox", ".sidebar", ".thumbcaption",
	}, ",")).Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// 3) Extract paragraphs (clean text blocks)
	parts := make([]string, 0, 64)
	root.Find("p").Each(func(i int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if t == "" {
			return
		}
		// skip tiny fragments
		if len([]rune(t)) < 40 {
			return
		}
		parts = append(parts, t)
	})

	// 4) Fallback if no paragraphs were found
	if len(parts) == 0 {
		t := cleanText(root.Text())
		if t == "" {
			return title, "", nil
		}
		return title, t, nil
	}

	return title, strings.Join(parts, "\n\n"), nil
}

func cleanText(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return ""
	}
	return strings.Join(strings.Fields(in), " ")
}
