package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	GitHubToken  string
	Repositories []struct {
		Owner string
		Name  string
	}
	FetchDir  string
	OutputDir string
	SiteRoot  string
	Since     time.Time
}

func main() {
	// Define command-line flags
	fetchFlag := flag.Bool("fetch", false, "Fetch GitHub activity data")
	renderFlag := flag.Bool("render", false, "Render static site from fetched data")
	serveFlag := flag.Bool("serve", false, "Serve render output dir")
	siteRootFlag := flag.String("site-root", "/", "Site root for render")
	flag.Parse()

	// account for the weekend when fetching two days back
	now := time.Now()
	workdaysFound := 0
	daysBack := 0

	for workdaysFound < 2 {
		daysBack++
		checkDate := now.AddDate(0, 0, -daysBack)
		weekday := checkDate.Weekday()

		// Skip Saturday (6) and Sunday (0)
		if weekday != time.Saturday && weekday != time.Sunday {
			workdaysFound++
		}
	}

	// Load configuration
	config := Config{
		GitHubToken: os.Getenv("KOKKOS_DASHBOARD_TOKEN"),
		Repositories: []struct {
			Owner string
			Name  string
		}{
			{Owner: "kokkos", Name: "kokkos"},
			{Owner: "kokkos", Name: "kokkos-kernels"},
			{Owner: "kokkos", Name: "kokkos-comm"},
			{Owner: "kokkos", Name: "kokkos-fft"},
			{Owner: "kokkos", Name: "kokkos-core-wiki"},
		},
		FetchDir:  "data/",
		OutputDir: "public/",
		SiteRoot:  *siteRootFlag,
		Since:     now.AddDate(0, 0, -daysBack),
	}

	// fetch
	if *fetchFlag {
		fmt.Println("Fetching GitHub activity...")
		fetch(config)
	}

	// render
	if *renderFlag {
		fmt.Println("Rendering static site...")
		render(config)
	}

	if *serveFlag {
		fs := http.FileServer(http.Dir(config.OutputDir))
		http.Handle("/", fs)
		log.Println("serve on :8080")
		http.ListenAndServe(":8080", nil)
	}
}
