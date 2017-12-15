package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	http           string
	token          string
	url            string
	refreshMinutes int
)

func init() {
	flag.StringVar(&http, "http", ":6060", "HTTP service address (e.g., ':6060')")
	flag.StringVar(&token, "gh-token", "", "A Github access token")
	flag.StringVar(&url, "gh-url", "", "The URL to Github Enterprise")
	flag.IntVar(&refreshMinutes, "refresh-minutes", 60,
		"The number of minutes to to wait in between each refresh")
	flag.Parse()

	if token == "" {
		fmt.Println("A github access token is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if url == "" {
		fmt.Println("The Github Enterprise URL is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	client, err := github.NewEnterpriseClient(url, url, oauth2.NewClient(ctx, ts))
	if err != nil {
		log.Fatalf("failed to initialize the Github client: %v", err)
	}

	// Do an initial refresh from GitHub
	log.Print("Starting initial refresh of all repositories")
	err = gogetAll(ctx, client)
	if err != nil {
		log.Fatalf("Initial refresh of repositories failed: %v", err)
	}

	ticker := time.NewTicker(time.Minute * time.Duration(refreshMinutes))
	go func() {
		for range ticker.C {
			log.Print("Beginning a refresh of all repositories")
			err = gogetAll(ctx, client)
			if err != nil {
				log.Printf("failed to download repositories from GitHub: %v", err)
			}
			log.Print("Refresh complete")
		}
	}()

	log.Print("Starting godoc")

	// serve documentation over godoc
	godoc := exec.Command("godoc", fmt.Sprintf("-http=%s", http)) // nolint: gas
	err = godoc.Run()

	log.Printf("godoc exited with error: %v", err)
}

func gogetAll(ctx context.Context, cl *github.Client) error {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	var allRepos []github.Repository
	for {
		res, resp, err := cl.Search.Repositories(ctx, "language:go", opts)
		if err != nil {
			return fmt.Errorf("failed to search GitHub: %v", err)
		}
		allRepos = append(allRepos, res.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// `go get` each repo
	for _, repo := range allRepos {
		if repo.GitURL == nil {
			continue
		}

		repoURL := *repo.GitURL
		repoURL = strings.TrimPrefix(repoURL, "git://")

		goget := exec.Command("go", "get", "-d", repoURL) // nolint: gas
		out, err := goget.CombinedOutput()
		if err != nil {
			log.Printf("Failed to get the repository: %s\n%s", repoURL, out)
		}
	}

	return nil
}
