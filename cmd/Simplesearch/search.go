package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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
		loadEnv(".env")

		// Support -q flag
		finalQuery := queryFlag
		if len(args) > 0 && finalQuery == "" {
			finalQuery = args[0]
		}

		runTimeout := getEnvAsDuration("RUN_TIMEOUT", 2*time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
		defer cancel()

		dbPath := getEnv("DB_PATH", "data/Simplesearch.db")
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

		// Reindex logic (Triggered internally if DB is empty, as requested)
		if idx == nil || idx.DocCount() == 0 {
			var seedURL string
			defaultSeedURL := getEnv("DEFAULT_SEED_URL", "")

			if finalQuery != "" {
				seedQuery := finalQuery
				if corrector != nil {
					if corrected, source := corrector.Correct(seedQuery); corrected != "" && corrected != seedQuery {
						log.Printf("auto-correct seed query (%s): %q -> %q", source, seedQuery, corrected)
						seedQuery = corrected
					}
				}

				seedURL, err = seed.ResolveWikipediaSeed(seedQuery)
				if err != nil {
					log.Printf("query seed failed, using default seed: %v", err)
					seedURL = defaultSeedURL
				}
			} else {
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
			Query: finalQuery,
			TopK:  topKFlag,
		})
		if err != nil {
			log.Fatalf("search failed: %v", err)
		}

		// Smarter Fallback mechanism: check if we actually found a relevant topic
		needsRecrawl := len(resp.Results) == 0

		if !needsRecrawl {
			// If we found local results, check if the query is actually in any of the titles.
			// in this case if same word comes in other articles we need a optional
			//check fallback 
			//mechnaism
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
			seedQuery := finalQuery
			if corrector != nil {
				if corrected, source := corrector.Correct(seedQuery); corrected != "" && corrected != seedQuery {
					log.Printf("auto-correct seed query (%s): %q -> %q", source, seedQuery, corrected)
					seedQuery = corrected
				}
			}

			seedURL, err := seed.ResolveWikipediaSeed(seedQuery)
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
				Query: finalQuery,
				TopK:  topKFlag,
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
// Helper Functions (PRESERVED EXACTLY AS SAME AS NEEDED MY MAIN.GO)
// ---------------------------------------------------------

func resolveRedisURL(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	return getEnv("REDIS_URL", "")
}

func crawlAndBuildIndex(ctx context.Context, seedURL string) (*index.Index, error) {
	maxPages := getEnvAsInt("MAX_PAGES", 50)
	maxTotalBytes := getEnvAsInt("MAX_TOTAL_BYTES", 5*1024*1024)
	maxBytesPerPage := getEnvAsInt("MAX_BYTES_PER_PAGE", 512*1024)
	workerCount := getEnvAsInt("WORKER_COUNT", 4)
	maxDepthInclusive := getEnvAsInt("MAX_DEPTH_INCLUSIVE", 3)
	userAgent := getEnv("USER_AGENT", "")

	cfg := crawler.Config{
		SeedURL:           seedURL,
		MaxPages:          maxPages,
		MaxTotalBytes:     int64(maxTotalBytes),
		MaxBytesPerPage:   int64(maxBytesPerPage),
		Workers:           workerCount,
		UserAgent:         userAgent,
		MaxDepthInclusive: maxDepthInclusive,
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

func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf(".env file not found: %s. Using defaults.", path)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}