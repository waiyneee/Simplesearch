package cli


import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

//a root.go tells env vars loads configs as well .
var rootCmd = &cobra.Command{
	Use:   "simplesearch",
	Short: "A fast, local Wikipedia search engine",
	Long:  `Simplesearch crawls Wikipedia and provides fast, fuzzy search capabilities 
	directly from your terminal.`,
	// If a user types just 'simplesearch', Cobra automatically prints the help menu.
	//pretty much detailed 
}


func Execute(){
	loadEnv(".env")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}