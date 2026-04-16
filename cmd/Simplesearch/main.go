package main

import (
	"context"
	"log"
	"time"

	"github.com/waiyneee/Simplesearch/internal/crawler"
)

func main() {
	// Safety timeout for full crawl run.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	stats, err := crawler.Run(ctx, crawler.Config{
		SeedURL:           "https://en.wikipedia.org/wiki/Cristiano_Ronaldo",
		MaxPages:          50,
		MaxTotalBytes:     5 * 1024 * 1024, // 5 MB
		MaxBytesPerPage:   512 * 1024,      // 512 KB/page
		Workers:           4,
		UserAgent:         "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)",
		MaxDepthInclusive: 3,
	})
	if err != nil {
		log.Fatalf("crawl failed: %v (scheduled=%d success=%d failed=%d bytes=%d)",
			err, stats.Scheduled, stats.Successful, stats.Failed, stats.TotalBytes)
	}

	log.Printf("crawl success: scheduled=%d success=%d failed=%d bytes=%d duration=%s",
		stats.Scheduled, stats.Successful, stats.Failed, stats.TotalBytes,
		stats.FinishedAt.Sub(stats.StartedAt))
}