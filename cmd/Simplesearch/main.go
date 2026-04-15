package main

import (
	"log"

	"github.com/waiyneee/Simplesearch/internal/crawler"
)

func main() {
	if err := crawler.Run(); err != nil {
		log.Fatal(err)
	}
}