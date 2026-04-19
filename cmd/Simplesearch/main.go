package main

import (
	"context"
	"log"
	"time"

	"github.com/waiyneee/Simplesearch/internal/crawler"
	"github.com/waiyneee/Simplesearch/internal/index"
	"github.com/waiyneee/Simplesearch/internal/pipeline"
)

const (
	runTimeout        = 2 * time.Minute
	seedURL           = "https://en.wikipedia.org/wiki/Cristiano_Ronaldo"
	maxPages          = 50
	maxTotalBytes     = 5 * 1024 * 1024 // 5 MB
	maxBytesPerPage   = 512 * 1024      // 512 KB
	workerCount       = 4
	maxDepthInclusive = 3
	userAgent         = "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	cfg := crawler.Config{
		SeedURL:           seedURL,
		MaxPages:          maxPages,
		MaxTotalBytes:     maxTotalBytes,
		MaxBytesPerPage:   maxBytesPerPage,
		Workers:           workerCount,
		UserAgent:         userAgent,
		MaxDepthInclusive: maxDepthInclusive,
	}

	crawlStats, pages, err := crawler.Run(ctx, cfg)
	if err != nil {
		log.Fatalf(
			"crawl failed: err=%v scheduled=%d successful=%d failed=%d bytes=%d duration=%s",
			err,
			crawlStats.Scheduled,
			crawlStats.Successful,
			crawlStats.Failed,
			crawlStats.TotalBytes,
			crawlStats.FinishedAt.Sub(crawlStats.StartedAt),
		)
	}

	idx := index.New()

	indexed, duplicates, indexErrs := 0, 0, 0
	for _, p := range pages {
		out := pipeline.IndexPage(idx, pipeline.PageToIndex{
			URL:   p.URL,
			Title: p.Title,
			Body:  p.BodyText,
		})

		if out.Err != nil {
			indexErrs++
			continue
		}
		if out.Added {
			indexed++
		} else {
			duplicates++
		}
	}

	log.Printf(
		"crawl completed: scheduled=%d successful=%d failed=%d bytes=%d duration=%s",
		crawlStats.Scheduled,
		crawlStats.Successful,
		crawlStats.Failed,
		crawlStats.TotalBytes,
		crawlStats.FinishedAt.Sub(crawlStats.StartedAt),
	)

	log.Printf(
		"index completed: indexed=%d duplicates=%d index_errs=%d pages_from_crawl=%d",
		indexed, duplicates, indexErrs, len(pages),
	)
}
