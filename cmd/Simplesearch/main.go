package main

import (
	"log"

	"github.com/waiyneee/Simplesearch/internal/crawler"
)

func main() {
	err := crawler.Run()
	if err != nil {
		log.Println("An error occured while crawling the data")
		log.Fatal(err)
	}
}