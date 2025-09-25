package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
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
}

func main() {
	// Define command-line flags
	fetchFlag := flag.Bool("fetch", false, "Fetch GitHub activity data")
	renderFlag := flag.Bool("render", false, "Render static site from fetched data")
	serveFlag := flag.Bool("serve", false, "Serve render output dir")
	siteRootFlag := flag.String("site-root", "/", "Site root for render")
	flag.Parse()

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
		http.ListenAndServe(":8080", nil)
	}
}
