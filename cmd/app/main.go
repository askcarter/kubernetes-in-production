package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	var (
		httpAddr = flag.String("http", ":80", "HTTP service address.")
	)
	flag.Parse()

	a := &app{}

	// Use a buffered error channel so that handlers can
	// keep processing after throwing errors.
	errChan := make(chan error, 10)
	go func() {
		httpServer := new(http.Server)
		httpServer.Addr = *httpAddr

		r := router(a)
		httpServer.Handler = loggingHandler(r)

		log.Println("Starting server...")
		log.Printf("HTTP service listening on %s", *httpAddr)

		errChan <- httpServer.ListenAndServe()
	}()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			// Log any errors from our server
			log.Fatal(err)
		case s := <-signalChan:
			// ctrl+c is a clean exit
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			os.Exit(0)
		}
	}
}

func router(a *app) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/heavy-workload", appHandler(a.heavy)).Methods("GET")
	return r
}

type app struct {
	// ds db.DataSource
}

func (a *app) heavy(w http.ResponseWriter, r *http.Request) (int, error) {
	x := 0.0001
	for i := 0; i <= 1000000; i++ {
		x += math.Sqrt(x)
	}

	fmt.Fprintf(w, `{"message": "Finished heavy computation"}`)
	return http.StatusOK, nil
}

// appHandler server all of this applications web traffic, handling
// error reporting and any setup that might be needed for our requests.
type appHandler func(http.ResponseWriter, *http.Request) (int, error)

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	go func(w http.ResponseWriter, r *http.Request) {
		status, err := fn(w, r)
		if err != nil {
			log.Println(err)
		}
		if status >= 400 {
			http.Error(w, http.StatusText(status), status)
		}
	}(w, r)
}

func loggingHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := "%s - - [%s] \"%s %s %s\" %s\n"
		fmt.Printf(format, r.RemoteAddr, time.Now().Format(time.RFC1123),
			r.Method, r.RequestURI, r.Proto, r.UserAgent())
		h.ServeHTTP(w, r)
	}
}
