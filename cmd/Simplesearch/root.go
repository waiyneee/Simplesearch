package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global flag variables
var (
	queryFlag     string
	topKFlag      int
	bodyLinesFlag int
	wrapWidthFlag int
	cacheModeFlag string
	redisURLFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "Simplesearch",
	Short: "A simple CLI app for managing search responses over Wikipedia",
	Long:  "An advanced fuzzy searching CLI tool that uses Wikipedia's pages to crawl content and indexes it.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Simplesearch! Run 'Simplesearch search -q \"your query\"' to start.")
	},
}

func Execute() {
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(suggestCmd)

	// Use native OS env vars for defaults if they exist
	//no need of loadEnv here
	cacheDef := os.Getenv("CACHE_MODE")
	if cacheDef == "" {
		cacheDef = "redis" //redis would be installed
	}

	redisDef := os.Getenv("REDIS_URL")
	if redisDef == "" {
		redisDef = "redis://localhost:6379"
	}
	// Map flags to Cobra
	rootCmd.PersistentFlags().StringVarP(&queryFlag, "query", "q", "", "search query")
	rootCmd.PersistentFlags().IntVarP(&topKFlag, "limit", "k", 10, "number of results to return")
	rootCmd.PersistentFlags().IntVar(&bodyLinesFlag, "body-lines", 8, "max lines of snippet to show per result")
	rootCmd.PersistentFlags().IntVar(&wrapWidthFlag, "wrap", 110, "wrap width for snippet output")

	rootCmd.PersistentFlags().StringVar(&cacheModeFlag, "cache", cacheDef, "cache mode: memory|redis")
	rootCmd.PersistentFlags().StringVar(&redisURLFlag, "redis-url", redisDef, "redis connection URL (overrides REDIS_URL)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
