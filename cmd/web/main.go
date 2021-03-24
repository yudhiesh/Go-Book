package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
)

// Configurations struct that stores all the flags that can be passed when
// running the server
type Config struct {
	Addr      string
	StaticDir string
}

// Define an application struct to hold the application-wide dependencies for
// the web application.
// These fields will be inherited by the handler methods that need the same
// logger functionality passed to them
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	// Define a new command-line flag with the name 'addr', default value of
	// 4000 for the port, and short explanation of the flag
	// Converts whatever value you pass into a string
	// addr := flag.String("addr", ":4000", "HTTP network address")

	// This parses the command-line flag.
	// This needs to be called before using the flag variables such as addr
	cfg := new(Config)
	// As the strings are stored in a struct we can access them using
	// flag.StringVar() instead of flag.String()

	flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.StaticDir, "static-dir", "./ui/static", "Path to static assets")

	flag.Parse()

	// the destination to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Creating a new logger with error information stderr as
	// the destination and use the log.Lshortfile flag to include the relevant
	// file name and line number.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	mux := http.NewServeMux()

	// HandleFunc takes in normal functions that are not actually Handlers as
	// they do not have the method ServeHTTP
	mux.HandleFunc("/", app.home)
	// NOTE: If you wanted to turn home into an actual handler you would need to
	// instantiate an interface home and then turn the home function into a
	// ServeHTTP method
	// Then pass it by pointer as below:
	// mux.Handle("/", &home{})
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)

	fileServer := http.FileServer(http.Dir(cfg.StaticDir))

	mux.Handle("/static/", http.StripPrefix("/static", neuter(fileServer)))

	// Override the http.Server Error Log
	// By default if Go's HTTP server encounters an error it will log it using
	// the standard logger
	// By initializing a new http.Server struct with the config settings of the
	// current server we can override it to use the errorLog
	srv := &http.Server{
		Addr:     cfg.Addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("Starting server on %s", cfg.Addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
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
