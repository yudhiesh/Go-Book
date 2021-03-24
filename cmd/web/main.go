package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

type Config struct {
	Addr      string
	StaticDir string
}

func main() {
	// Define a new command-line flag with the name 'addr', default value of
	// 4000 for the port, and short explanation of the flag
	// Converts whatever value you pass into a string
	// addr := flag.String("addr", ":4000", "HTTP network address")

	// This parses the command-line flag.
	// This needs to be called before using the flag variables such as addr
	cfg := new(Config)
	flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.StaticDir, "static-dir", "./ui/static", "Path to static assets")
	flag.Parse()

	mux := http.NewServeMux()

	// HandleFunc takes in normal functions that are not actually Handlers as
	// they do not have the method ServeHTTP
	mux.HandleFunc("/", home)
	// NOTE: If you wanted to turn home into an actual handler you would need to
	// instantiate an interface home and then turn the home function into a
	// ServeHTTP method
	// Then pass it by pointer as below:
	// mux.Handle("/", &home{})
	mux.HandleFunc("/snippet", showSnippet)
	mux.HandleFunc("/snippet/create", createSnippet)

	fileServer := http.FileServer(http.Dir(cfg.StaticDir))

	mux.Handle("/static/", http.StripPrefix("/static", neuter(fileServer)))

	log.Printf("Starting server on %s", cfg.Addr)
	err := http.ListenAndServe(cfg.Addr, mux)
	log.Fatal(err)
}

// Disable http.FileServer Directory Listings
// NOTE: Prevent navigable directory listings and disable it all together
func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
