package crawler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
)

type Config struct {
	SeedURL           string
	MaxPages          int   // max pages to schedule
	MaxTotalBytes     int64 // max total downloaded bytes across successful pages
	MaxBytesPerPage   int64
	Workers           int
	UserAgent         string
	MaxDepthInclusive int
}

type PageResult struct {
	URL        string
	Depth      int
	StatusCode int
	Bytes      int64
	Links      []string
	Title      string
	BodyText   string
	Err        error
}

type CrawlStats struct {
	Scheduled  int
	Successful int
	Failed     int
	TotalBytes int64
	StartedAt  time.Time
	FinishedAt time.Time
}

type job struct {
	URL   string
	Depth int
}

// Run starts crawl with worker pool + jobs/results channels.
// Frontier is owned by coordinator goroutine only (no mutex needed).
func Run(ctx context.Context, cfg Config) (CrawlStats, error) {
	var stats CrawlStats
	stats.StartedAt = time.Now()

	if cfg.SeedURL == "" {
		return stats, fmt.Errorf("seed URL is required")
	}
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 50
	}
	if cfg.MaxTotalBytes <= 0 {
		cfg.MaxTotalBytes = 5 * 1024 * 1024 // 5 MB
	}
	if cfg.MaxBytesPerPage <= 0 {
		cfg.MaxBytesPerPage = 512 * 1024 // 512 KB/page
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}
	if cfg.MaxDepthInclusive <= 0 {
		cfg.MaxDepthInclusive = 3
	}

	seedParsed, err := url.Parse(cfg.SeedURL)
	if err != nil {
		return stats, fmt.Errorf("invalid seed URL: %w", err)
	}
	if seedParsed.Host != "en.wikipedia.org" {
		return stats, fmt.Errorf("seed must be en.wikipedia.org")
	}

	frontier := NewFrontier()
	frontier.Add(cfg.SeedURL, 0)

	jobs := make(chan job)
	results := make(chan PageResult, cfg.Workers*2)

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			body, status, ferr := Fetch(ctx, j.URL, cfg.UserAgent, cfg.MaxBytesPerPage)
			if ferr != nil {
				results <- PageResult{
					URL:   j.URL,
					Depth: j.Depth,
					Err:   ferr,
				}
				continue
			}

			base, perr := url.Parse(j.URL)
			if perr != nil {
				results <- PageResult{
					URL:        j.URL,
					Depth:      j.Depth,
					StatusCode: status,
					Bytes:      int64(len(body)),
					Err:        perr,
				}
				continue
			}

			links, lerr := ExtractLinks(body, base)
			title, bodyText, cerr := ExtractPageContent(body)

			var combinedErr error
			if lerr != nil {
				combinedErr = lerr
			}
			if cerr != nil && combinedErr == nil {
				combinedErr = cerr
			}

			results <- PageResult{
				URL:        j.URL,
				Depth:      j.Depth,
				StatusCode: status,
				Bytes:      int64(len(body)),
				Links:      links,
				Title:      title,
				BodyText:   bodyText,
				Err:        combinedErr,
			}
		}
	}

	wg.Add(cfg.Workers)
	for i := 0; i < cfg.Workers; i++ {
		go worker()
	}

	inFlight := 0

	// schedule pushes one frontier item to jobs if allowed by limits.
	schedule := func(item FrontierItem) bool {
		if item.Depth > cfg.MaxDepthInclusive {
			return false
		}
		if stats.Scheduled >= cfg.MaxPages {
			return false
		}
		if stats.TotalBytes >= cfg.MaxTotalBytes {
			return false
		}

		select {
		case <-ctx.Done():
			return false
		case jobs <- job{URL: item.URL, Depth: item.Depth}:
			stats.Scheduled++
			inFlight++
			return true
		}
	}

	// Prime workers initially.
	for inFlight < cfg.Workers {
		item, ok := frontier.Next()
		if !ok {
			break
		}
		_ = schedule(item)
	}

	// Coordinator loop
	for inFlight > 0 {
		select {
		case <-ctx.Done():
			// Stop scheduling new work; drain in-flight results.
			for inFlight > 0 {
				<-results
				inFlight--
			}
			close(jobs)
			wg.Wait()
			close(results)
			stats.FinishedAt = time.Now()
			return stats, ctx.Err()

		case res := <-results:
			inFlight--

			if res.Err != nil {
				stats.Failed++
				log.Printf("crawl error url=%s depth=%d err=%v", res.URL, res.Depth, res.Err)
			} else {
				stats.Successful++
				stats.TotalBytes += res.Bytes

				log.Printf("ok scheduled=%d success=%d failed=%d depth=%d status=%d bytes=%d url=%s links=%d title=%q body_chars=%d",
					stats.Scheduled, stats.Successful, stats.Failed,
					res.Depth, res.StatusCode, res.Bytes, res.URL, len(res.Links), res.Title, len(res.BodyText))

				if res.Depth < cfg.MaxDepthInclusive {
					nextDepth := res.Depth + 1
					for _, link := range res.Links {
						frontier.Add(link, nextDepth)
					}
				}
			}

			reachedPages := stats.Scheduled >= cfg.MaxPages
			reachedBytes := stats.TotalBytes >= cfg.MaxTotalBytes

			if reachedPages || reachedBytes {
				// No more scheduling; drain existing in-flight jobs.
				continue
			}

			// Keep workers busy up to configured count.
			for inFlight < cfg.Workers {
				item, ok := frontier.Next()
				if !ok {
					break
				}
				if !schedule(item) {
					break
				}
			}
		}
	}

	close(jobs)
	wg.Wait()
	close(results)

	stats.FinishedAt = time.Now()
	log.Printf("crawl finished scheduled=%d success=%d failed=%d total_bytes=%d duration=%s frontier_remaining=%d",
		stats.Scheduled, stats.Successful, stats.Failed, stats.TotalBytes,
		stats.FinishedAt.Sub(stats.StartedAt), frontier.Len())

	return stats, nil
}