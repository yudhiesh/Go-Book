package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"yudhiesh/snippetbox/pkg/models"
	"yudhiesh/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

// Custom context key
type contextKey string

// Context key is authenticated variable
const contextKeyIsAuthenticated = contextKey("isAuthenticated")

// Define an Application struct to hold the Application-wide dependencies for
// the web Application.
// These fields will be inherited by the handler methods that need the same
// logger functionality passed to them
type Application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	session  *sessions.Session
	// Make snippets and user take in generic types/interfaces instead of concrete types of
	// *mysql.SnippetMode and *mysql.UserModel
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
	}
}

func main() {
	// This parses the command-line flag.
	// This needs to be called before using the flag variables such as addr
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:password@/snippetbox?parseTime=true", "MySQL data source name")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

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

	// Connect to the DB
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Closes the connection pool before the main function finishes
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize a new session manager with the secret key
	// It is configured to always expires after 12 hours
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	app := &Application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
		users:         &mysql.UserModel{DB: db},
	}

	// tls.Config struct holds the non-default TLS setting we want the server to
	// use
	tlsConfig := &tls.Config{
		// uses Go's favored cipher suites or the users preferred cipher suite
		// which will probably have a stronger cipher suite
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Override the http.Server Error Log
	// By default if Go's HTTP server encounters an error it will log it using
	// the standard logger
	// By initializing a new http.Server struct with the config settings of the
	// current server we can override it to use the errorLog
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(),
		TLSConfig: tlsConfig,
		// keep-alives last a minute instead of 3 minutes
		IdleTimeout: time.Minute,
		// Short ReadTimeout prevents SLowloris attacks
		ReadTimeout: 5 * time.Second,
		// Prevent the data that the handler returns from taking too long to
		// write
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	// ListenAndServeTLS() is used to start the HTTPS server
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// Returns a sql.DB connection pool for a given DSN
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Connections are established lazily as and when needed for the first time
	// db.Ping creates a connection and we check that there isn't any errors
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
