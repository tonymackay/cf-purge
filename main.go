package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

var (
	showVersion bool
	version     = "dev"
	dryRun      bool
	url         string
	file        string
	urls        = make(map[string]struct{})
	apiToken    = os.Getenv("CF_API_TOKEN")
	zoneID      = os.Getenv("CF_ZONE_ID")
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "print version number")
	flag.BoolVar(&dryRun, "dryrun", false, "run command without making changes")
	flag.StringVar(&url, "url", "", "purge a single URL from Cloudflare's cache (eg https://example.com)")
	flag.StringVar(&file, "file", "", "purge multiple URLs from Cloudflare's cache")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (runtime: %s)\n", os.Args[0], version, runtime.Version())
		os.Exit(0)
	}

	if apiToken == "" || zoneID == "" {
		fmt.Println("error: CF_API_TOKEN and CF_ZONE_ID are required")
		flag.Usage()
		os.Exit(2)
	}

	if url == "" && file == "" {
		fmt.Println("error: -url or -file option is required")
		flag.Usage()
		os.Exit(2)
	}

	if url != "" {
		urls[url] = struct{}{}
	}

	loadURLSfromFile()
	purge()
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cf-purge [OPTIONS]")
	fmt.Fprintln(os.Stderr, "\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "ENVIRONMENT:")
	fmt.Fprintln(os.Stderr, "  CF_API_TOKEN    the Cloudflare API Token with Zone.Cache Purge permission")
	fmt.Fprintln(os.Stderr, "  CF_ZONE_ID      the Cloudflare Zone ID of the domain to purge")
	fmt.Fprintln(os.Stderr, "\nEXAMPLE:")
	fmt.Fprintln(os.Stderr, "  export CF_API_TOKEN=<token> \n  export CF_ZONE_ID=<zone>")
	fmt.Fprintln(os.Stderr, "  cf-purge -url https://example.com or cf-purge -file urls.txt")
	fmt.Fprintln(os.Stderr, "")
}

func loadURLSfromFile() {
	if file == "" {
		return
	}

	file, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls[strings.TrimSpace(scanner.Text())] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func purge() {
	fmt.Println("Purging URLs from Cloudflare's edge cache")
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		log.Fatal(err)
	}
	// purge URLS in batches of 30
	chunkSize := 30
	batchKeys := make([]string, 0, chunkSize)
	process := func() {
		time.Sleep(time.Second)
		if dryRun == false {
			pcr := cloudflare.PurgeCacheRequest{Files: batchKeys}
			r, err := api.PurgeCache(zoneID, pcr)
			if err != nil {
				log.Fatal(err)
			}
			if r.Success {
				fmt.Println("purged: ", batchKeys)
			} else {
				fmt.Println("purge failed: ", batchKeys)
			}
		} else {
			fmt.Println("(dryrun) purged: ", batchKeys)
		}
		batchKeys = batchKeys[:0]
	}
	// create each batch of URLS and run the purge
	for k := range urls {
		batchKeys = append(batchKeys, k)
		if len(batchKeys) == chunkSize {
			process()
		}
	}
	// Process last, potentially incomplete batch
	if len(batchKeys) > 0 {
		process()
	}
}
