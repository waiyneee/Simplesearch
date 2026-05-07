package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/waiyneee/Simplesearch/internal/app"
	"github.com/waiyneee/Simplesearch/internal/format"
	"github.com/waiyneee/Simplesearch/internal/storage"
)

var (
	topK      int
	bodyLines int
	wrapWidth int
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the indexed Wikipedia pages",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		
		runTimeout := getEnvAsDuration("RUN_TIMEOUT", 2*time.Minute)
		_, cancel := context.WithTimeout(context.Background(), runTimeout)
		defer cancel()

		dbPath := getEnv("DB_PATH", "data/Simplesearch.db")
		db, err := storage.OpenDbInstance(dbPath)
		if err != nil {
			log.Fatalf("open db failed: %v", err)
		}
		defer db.Close()

		idx, err := storage.LoadIndex(db)
		if err != nil || idx == nil || idx.DocCount() == 0 {
			log.Fatalf("Index is empty. Run 'simplesearch index' first.")
		}

		// Notice how cache is treated as background business logic, not a flag here.
		cacheMode := getEnv("CACHE_MODE", "memory")
		redisURL := getEnv("REDIS_URL", "")

		application, err := app.New(idx, app.Config{
			CacheMode: cacheMode,
			RedisURL:  redisURL,
		})
		if err != nil {
			log.Fatalf("app init failed: %v", err)
		}

		resp, err := application.Run(app.SearchRequest{
			Query: query,
			TopK:  topK,
		})
		if err != nil {
			log.Fatalf("search failed: %v", err)
		}

		if len(resp.Results) == 0 {
			fmt.Println("No results found.")
			return
		}

		fmt.Printf("\nTop %d results for query: %q\n\n", len(resp.Results), query)
		for i, r := range resp.Results {
			fmt.Printf("%d) %s\n", i+1, r.Title)
			fmt.Printf("   URL: %s\n", r.URL)
			snippet := format.WrapText(r.Snippet, wrapWidth)
			fmt.Printf("   %s\n\n", format.TruncateLines(snippet, bodyLines))
		}
		_ = os.Stdout.Sync()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVarP(&topK, "limit", "k", 10, "number of results to return")
	searchCmd.Flags().IntVar(&bodyLines, "body-lines", 8, "max lines of snippet")
	searchCmd.Flags().IntVar(&wrapWidth, "wrap", 110, "wrap width for output")
}