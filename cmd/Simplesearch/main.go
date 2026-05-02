package main

import (
	"context"
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
	"github.com/waiyneee/Simplesearch/internal/storage"
	"github.com/waiyneee/Simplesearch/internal/seed"
)

const (
	runTimeout        = 2 * time.Minute
	defaultSeedURL    = "https://en.wikipedia.org/wiki/Cristiano_Ronaldo" //as a fallback
	maxPages          = 50
	maxTotalBytes     = 5 * 1024 * 1024 // 5 MB
	maxBytesPerPage   = 512 * 1024      // 512 KB
	workerCount       = 4
	maxDepthInclusive = 3
	userAgent         = "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)"

	dbPath = "data/Simplesearch.db"
)

func main() {
	query := flag.String("q", "", "search query")
	topKvalue := flag.Int("k", 10, "number of results to return")
	reindex := flag.Bool("reindex", false, "force fresh crawl+index and overwrite DB")
	bodyLines := flag.Int("body-lines", 8, "max lines of snippet to show per result")
	wrapWidth := flag.Int("wrap", 110, "wrap width for snippet output")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	db, err := storage.OpenDbInstance(dbPath)
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}
	defer db.Close()

	if err := storage.CreateSchema(db); err != nil {
		log.Fatalf("create schema failed: %v", err)
	}

	var idx *index.Index

	// Try to load existing index unless user forces reindex.
	if !*reindex {
		idx, err = storage.LoadIndex(db)
		if err != nil {
			log.Printf("load index failed: %v", err)
		}
	}

	// If no index (or forced reindex), build from query seed first.
	if *reindex || idx == nil || idx.DocCount() == 0 {
		var seedURL string

		// Prefer query-based seed when query is provided.
		if *query != "" {
			seedURL, err = seed.ResolveWikipediaSeed(*query)
			if err != nil {
				log.Printf("query seed failed, using default seed: %v", err)
				seedURL = defaultSeedURL
			}
		} else {
			// No query provided → use fallback seed.
			seedURL = defaultSeedURL
		}

		log.Printf("building fresh index from seed: %s", seedURL)

		idx, err = crawlAndBuildIndex(ctx, seedURL)
		if err != nil {
			log.Fatalf("crawl/index failed: %v", err)
		}
		if err := storage.SaveIndex(db, idx); err != nil {
			log.Fatalf("save index failed: %v", err)
		}
	}

	if idx == nil {
		log.Fatalf("index is nil after initialization")
	}

	if *query == "" {
		log.Printf("no query provided. run with -q \"your query\" to search")
		return
	}

	application, err := app.New(idx)
	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	resp, err := application.Run(app.SearchRequest{
		Query: *query,
		TopK:  *topKvalue,
	})
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	// If still no results, rebuild using query seed and try again.
	if len(resp.Results) == 0 {
		seedURL, err := seed.ResolveWikipediaSeed(*query)
		if err != nil {
			log.Fatalf("no results + failed to resolve seed: %v", err)
		}

		log.Printf("no results. crawling new seed: %s", seedURL)

		idx, err = crawlAndBuildIndex(ctx, seedURL)
		if err != nil {
			log.Fatalf("crawl/index failed: %v", err)
		}
		if err := storage.SaveIndex(db, idx); err != nil {
			log.Fatalf("save index failed: %v", err)
		}

		application, err = app.New(idx)
		if err != nil {
			log.Fatalf("app init failed after rebuild: %v", err)
		}

		resp, err = application.Run(app.SearchRequest{
			Query: *query,
			TopK:  *topKvalue,
		})
		if err != nil {
			log.Fatalf("search failed after rebuild: %v", err)
		}
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

		snippet := format.WrapText(r.Snippet, *wrapWidth)
		snippet = format.TruncateLines(snippet, *bodyLines)
		fmt.Printf("   %s\n\n", snippet)
	}

	_ = os.Stdout.Sync()
}

func crawlAndBuildIndex(ctx context.Context, seedURL string) (*index.Index, error) {
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