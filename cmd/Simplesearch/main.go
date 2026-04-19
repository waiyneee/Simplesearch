package main

import (
	"context"
	"log"
	"time"

	"github.com/waiyneee/Simplesearch/internal/crawler"
	"github.com/waiyneee/Simplesearch/internal/index"
)

const (
	runTimeout        = 2 * time.Minute
	seedURL           = "https://en.wikipedia.org/wiki/Cristiano_Ronaldo"
	maxPages          = 50
	maxTotalBytes     = 5 * 1024 * 1024 // 5 MB
	maxBytesPerPage   = 512 * 1024       // 512 KB
	workerCount       = 4
	maxDepthInclusive = 3
	userAgent         = "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	// Index is initialized now so crawl->index wiring can use it.
	// (Actual indexing call should happen where PageResult is available.)
	idx := index.New()
	_ = idx

	cfg := crawler.Config{
		SeedURL:           seedURL,
		MaxPages:          maxPages,
		MaxTotalBytes:     maxTotalBytes,
		MaxBytesPerPage:   maxBytesPerPage,
		Workers:           workerCount,
		UserAgent:         userAgent,
		MaxDepthInclusive: maxDepthInclusive,
	}

	stats, err := crawler.Run(ctx, cfg)
	if err != nil {
		log.Fatalf(
			"crawl failed: err=%v scheduled=%d successful=%d failed=%d bytes=%d duration=%s",
			err,
			stats.Scheduled,
			stats.Successful,
			stats.Failed,
			stats.TotalBytes,
			stats.FinishedAt.Sub(stats.StartedAt),
		)
	}

	log.Printf(
		"crawl completed: scheduled=%d successful=%d failed=%d bytes=%d duration=%s",
		stats.Scheduled,
		stats.Successful,
		stats.Failed,
		stats.TotalBytes,
		stats.FinishedAt.Sub(stats.StartedAt),
	)
}