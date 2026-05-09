package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/waiyneee/Simplesearch/internal/app"
	"github.com/waiyneee/Simplesearch/internal/crawler"
	"github.com/waiyneee/Simplesearch/internal/format"
	"github.com/waiyneee/Simplesearch/internal/index"
	"github.com/waiyneee/Simplesearch/internal/pipeline"
	"github.com/waiyneee/Simplesearch/internal/seed"
	"github.com/waiyneee/Simplesearch/internal/storage"
	"github.com/waiyneee/Simplesearch/internal/suggest"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search your query specifically",
	Long:  "We will auto fuzzy search and autocorrect your query to run our search indexing logic.",
	Run: func(cmd *cobra.Command, args []string) {
		// Support -q flag
		finalQuery := queryFlag
		if len(args) > 0 && finalQuery == "" {
			finalQuery = args[0]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 115*time.Second)
		defer cancel()

		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "data/Simplesearch.db"
		}

		db, err := storage.OpenDbInstance(dbPath)
		if err != nil {
			log.Fatalf("open db failed: %v", err)
		}
		defer db.Close()

		if err := storage.CreateSchema(db); err != nil {
			log.Fatalf("create schema failed: %v", err)
		}

		var idx *index.Index
		idx, err = storage.LoadIndex(db)
		if err != nil {
			log.Printf("load index failed: %v", err)
		}

		// build corrector (local + wiki)
		cache := suggest.NewMemoryCache()
		var local *suggest.LocalIndexSuggestor
		if idx != nil {
			local = suggest.NewLocalIndexSuggestor(idx)
		}
		wiki := suggest.NewWikiSuggestor()
		corrector := suggest.NewCorrector(cache, local, wiki)

		// Auto-correct the query ONCE upfront
		if finalQuery != "" && corrector != nil {
			if corrected, source := corrector.Correct(finalQuery); corrected != "" && corrected != finalQuery {
				log.Printf("auto-correct seed query (%s): %q -> %q", source, finalQuery, corrected)
				finalQuery = corrected
			}
		}

		// Reindex logic
		if idx == nil || idx.DocCount() == 0 {
			var seedURL string
			defaultSeedURL := os.Getenv("DEFAULT_SEED_URL")
			if defaultSeedURL == "" {
				defaultSeedURL = "https://en.wikipedia.org/wiki/Cristiano_Ronaldo"
			}

			if finalQuery != "" {
				seedURL, err = seed.ResolveWikipediaSeed(finalQuery)
				if err != nil {
					log.Printf("query seed failed, using default seed: %v", err)
					seedURL = defaultSeedURL
				}
			} else {
				seedURL = defaultSeedURL
			}

			if seedURL == "" {
				log.Fatalf("Index is empty and no valid query or DEFAULT_SEED_URL provided to seed it.")
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

		if finalQuery == "" {
			log.Printf("no query provided. run with -q \"your query\" or positional arg to search")
			return
		}

		redisURL := resolveRedisURL(redisURLFlag)

		application, err := app.New(idx, app.Config{
			CacheMode: cacheModeFlag,
			RedisURL:  redisURL,
		})
		if err != nil {
			log.Fatalf("app init failed: %v", err)
		}

		resp, err := application.Run(app.SearchRequest{
			Query:        finalQuery,
			TopK:         topKFlag,
			SnippetChars: bodyLinesFlag * wrapWidthFlag,
		})
		if err != nil {
			log.Fatalf("search failed: %v", err)
		}

		// Smarter Fallback mechanism
		needsRecrawl := len(resp.Results) == 0

		if !needsRecrawl {
			hasTitleMatch := false
			queryLower := strings.ToLower(strings.TrimSpace(finalQuery))
			for _, r := range resp.Results {
				if strings.Contains(strings.ToLower(r.Title), queryLower) {
					hasTitleMatch = true
					break
				}
			}
			if !hasTitleMatch {
				needsRecrawl = true
			}
		}

		if needsRecrawl {
			seedURL, err := seed.ResolveWikipediaSeed(finalQuery)
			if err != nil {
				log.Fatalf("no results + failed to resolve seed: %v", err)
			}

			log.Printf("local index lacks a dedicated article. crawling new seed: %s", seedURL)

			idx, err = crawlAndBuildIndex(ctx, seedURL)
			if err != nil {
				log.Fatalf("crawl/index failed: %v", err)
			}
			if err := storage.SaveIndex(db, idx); err != nil {
				log.Fatalf("save index failed: %v", err)
			}

			application, err = app.New(idx, app.Config{
				CacheMode: cacheModeFlag,
				RedisURL:  redisURL,
			})
			if err != nil {
				log.Fatalf("app init failed after rebuild: %v", err)
			}

			resp, err = application.Run(app.SearchRequest{
				Query:        finalQuery,
				TopK:         topKFlag,
				SnippetChars: bodyLinesFlag * wrapWidthFlag,
			})
			if err != nil {
				log.Fatalf("search failed after rebuild: %v", err)
			}
		}

		if len(resp.Results) == 0 {
			fmt.Println("No results found")
			return
		}

		fmt.Printf("\nTop %d results for query: %q\n\n", len(resp.Results), finalQuery)

		seen := make(map[string]struct{})
		rank := 1
		for _, r := range resp.Results {
			key := strings.ToLower(strings.TrimSpace(r.URL))
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			fmt.Printf("%d) %s\n", rank, r.Title)
			fmt.Printf("   URL: %s\n", r.URL)
			fmt.Printf("   Score: %.6f\n", r.Score)

			snippet := format.WrapText(r.Snippet, wrapWidthFlag)
			snippet = format.TruncateLines(snippet, bodyLinesFlag)
			fmt.Printf("   %s\n\n", snippet)

			rank++
		}

		_ = os.Stdout.Sync()
	},
}

// ---------------------------------------------------------
// Helper Functions (Cleaned up!)
// ---------------------------------------------------------

func resolveRedisURL(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	return os.Getenv("REDIS_URL")
}

func crawlAndBuildIndex(ctx context.Context, seedURL string) (*index.Index, error) {
	cfg := crawler.Config{
		SeedURL:           seedURL,
		MaxPages:          50,
		MaxTotalBytes:     5242890,
		MaxBytesPerPage:   524286,
		Workers:           4,
		UserAgent:         "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)",
		MaxDepthInclusive: 3,
	}

	crawlStats, pages, err := crawler.Run(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf(
			"crawl failed: err=%v scheduled=%d successful=%d failed=%d bytes=%d duration=%s",
			err, crawlStats.Scheduled, crawlStats.Successful, crawlStats.Failed, crawlStats.TotalBytes, crawlStats.FinishedAt.Sub(crawlStats.StartedAt),
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

	log.Printf("crawl completed: scheduled=%d successful=%d failed=%d bytes=%d duration=%s",
		crawlStats.Scheduled, crawlStats.Successful, crawlStats.Failed, crawlStats.TotalBytes, crawlStats.FinishedAt.Sub(crawlStats.StartedAt))

	log.Printf("index completed: indexed=%d duplicates=%d index_errs=%d pages_from_crawl=%d",
		indexed, duplicates, indexErrs, len(pages))

	return idx, nil
}
