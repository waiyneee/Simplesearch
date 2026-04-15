package crawler

import (
	"fmt"
	// "log"
	"net/url"
)

const seedURL = "https://en.wikipedia.org/wiki/Cristiano_Ronaldo"

func Run() error {
	body, resp, err := Fetch(seedURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))

	fmt.Println("\nFirst 500 chars of HTML:")
	if len(body) > 500 {
		fmt.Println(string(body[:500]))
	} else {
		fmt.Println(string(body))
	}

	base, err := url.Parse(seedURL)
	if err != nil {
		return err
	}

	links, err := ExtractLinks(body, base)
	if err != nil {
		return err
	}

	fmt.Println("\nExtracted Wikipedia links:")
	for i, link := range links {
		if i >= 20 {
			break
		}
		fmt.Println(link)
	}

	return nil
}