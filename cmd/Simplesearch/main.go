package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/waiyneee/Simplesearch/internal/app"
	"github.com/waiyneee/Simplesearch/internal/crawler"
	"github.com/waiyneee/Simplesearch/internal/format"
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

	snapshotPath = "data/index_snapshot.json"
)

func main() {
	query := flag.String("q", "", "search query")
	topKvalue := flag.Int("k", 10, "number of results to return")
	reindex := flag.Bool("reindex", false, "force fresh crawl+index and overwrite snapshot")
	bodyLines := flag.Int("body-lines", 8, "max lines of snippet to show per result")
	wrapWidth := flag.Int("wrap", 110, "wrap width for snippet output")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	var idx *index.Index
	var err error

	// 1) Try loading snapshot unless reindex is forced.
	if !*reindex {
		idx, err = index.Load(snapshotPath)
		switch {
		case err == nil:
			log.Printf("loaded index snapshot from %s", snapshotPath)

		case errors.Is(err, index.ErrSnapshotNotFound):
			log.Printf("snapshot not found at %s, running crawl+index", snapshotPath)
			idx, err = crawlAndBuildIndex(ctx)
			if err != nil {
				log.Fatalf("crawl/index failed: %v", err)
			}
			if err := idx.Save(snapshotPath); err != nil {
				log.Printf("warning: failed to save snapshot: %v", err)
			} else {
				log.Printf("saved index snapshot to %s", snapshotPath)
			}

		default:
			// Corrupt/unsupported/etc: warn and fallback.
			log.Printf("warning: failed to load snapshot (%v), running crawl+index fallback", err)
			idx, err = crawlAndBuildIndex(ctx)
			if err != nil {
				log.Fatalf("crawl/index fallback failed: %v", err)
			}
			if err := idx.Save(snapshotPath); err != nil {
				log.Printf("warning: failed to save snapshot after fallback: %v", err)
			} else {
				log.Printf("saved index snapshot to %s", snapshotPath)
			}
		}
	} else {
		// Forced reindex path
		log.Printf("reindex=true, skipping snapshot load and running fresh crawl+index")
		idx, err = crawlAndBuildIndex(ctx)
		if err != nil {
			log.Fatalf("crawl/index failed: %v", err)
		}
		if err := idx.Save(snapshotPath); err != nil {
			log.Printf("warning: failed to save snapshot: %v", err)
		} else {
			log.Printf("saved index snapshot to %s", snapshotPath)
		}
	}

	if idx == nil {
		log.Fatalf("index is nil after initialization")
	}

	// Optional searching
	if *query == "" {
		log.Printf("no query provided. run with -q \"your query input\" to search docs")
		return
	}

	application, err := app.New(idx)
	if err != nil {
		log.Fatalf("app initializer failed: %v", err)
	}

	resp, err := application.Run(app.SearchRequest{
		Query: *query,
		TopK:  *topKvalue,
	})
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	if len(resp.Results) == 0 {
		fmt.Println("No results found")
		return
	}

	fmt.Printf("\nTop %d results for query: %q\n\n", len(resp.Results), *query)
	for i, r := range resp.Results {
		fmt.Printf("%d) %s\n", i+1, r.Title)
		fmt.Printf("   URL: %s\n", r.URL)
		fmt.Printf("   Score: %.6f\n", r.Score)
		// fmt.Printf("   %s\n\n", r.Snippet)

		snippet := r.Snippet
		snippet = format.WrapText(snippet, *wrapWidth)
		snippet = format.TruncateLines(snippet, *bodyLines)

		fmt.Printf("   %s\n\n", snippet)

	}

	_ = os.Stdout.Sync()
}

func crawlAndBuildIndex(ctx context.Context) (*index.Index, error) {
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
		return nil, fmt.Errorf(
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

	return idx, nil
}
