package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/waiyneee/Simplesearch/internal/storage"
	"github.com/waiyneee/Simplesearch/internal/suggest"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest [query]",
	Short: "Get typo corrections and suggestions",
	Long:  "Provides spelling suggestions by querying your local offline index as well as the Wikipedia API.",
	Run: func(cmd *cobra.Command, args []string) {
		loadEnv(".env")

		finalQuery := queryFlag
		if len(args) > 0 && finalQuery == "" {
			finalQuery = args[0]
		}

		if finalQuery == "" {
			fmt.Println("Please provide a query to suggest corrections for. Example: simplesearch suggest \"appple\"")
			return
		}

		// 1. Connect to DB to load local index
		dbPath := getEnv("DB_PATH", "data/Simplesearch.db")
		db, err := storage.OpenDbInstance(dbPath)
		if err != nil {
			log.Fatalf("open db failed: %v", err)
		}
		defer db.Close()

		idx, err := storage.LoadIndex(db)
		if err != nil {
			log.Printf("Warning: Failed to load local index. Local suggestions may be unavailable. Error: %v", err)
		}

		fmt.Printf("Analyzing query: %q\n\n", finalQuery)

		// 2. Local Index Suggestion
		if idx != nil && idx.DocCount() > 0 {
			local := suggest.NewLocalIndexSuggestor(idx)
			// Using 2 lev distance and 0.35 trigram sim as set in your corrector.go defaults
			if val, ok := local.Suggest(finalQuery, 2, 0.35); ok && val != "" {
				fmt.Printf("✅ Local Index Suggestion: %s\n", val)
			} else {
				fmt.Println("❌ Local Index Suggestion: No close matches found locally.")
			}
		} else {
			fmt.Println("❌ Local Index Suggestion: Index is empty.")
		}

		// 3. Wiki API Suggestion
		wiki := suggest.NewWikiSuggestor()
		if val, err := wiki.TopTitle(finalQuery); err == nil && val != "" {
			fmt.Printf("🌍 Wikipedia API Suggestion: %s\n", val)
		} else {
			fmt.Println("🌍 Wikipedia API Suggestion: No suggestions found or API error.")
		}
	},
}