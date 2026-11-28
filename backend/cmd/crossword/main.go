// Command crossword is the single backend binary. It runs one of:
//
//	server            the HTTP API
//	scraper           the long-running scraper
//	scrapeall [DATE]  a one-off scrape of every date (from DATE, YYYY-MM-DD)
//
// The command is taken from the name the binary was invoked as (the image
// symlinks server.bin, scraper.bin and scrapeall.bin to it) or, failing that,
// from the first argument: `crossword server`. One binary instead of three
// shares a single copy of the Go runtime and dependencies.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/theodoretliu/crossword-leaderboard/internal/scrapeall"
	"github.com/theodoretliu/crossword-leaderboard/internal/scraper"
	"github.com/theodoretliu/crossword-leaderboard/internal/server"
)

var commands = map[string]func(args []string){
	"server":    func([]string) { server.Main() },
	"scraper":   func([]string) { scraper.Main() },
	"scrapeall": scrapeall.Main,
}

func main() {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".bin")
	args := os.Args[1:]
	if _, ok := commands[name]; !ok && len(args) > 0 {
		name, args = args[0], args[1:]
	}

	run, ok := commands[name]
	if !ok {
		fmt.Fprintln(os.Stderr, "usage: crossword server | scraper | scrapeall [YYYY-MM-DD]")
		os.Exit(2)
	}
	run(args)
}
